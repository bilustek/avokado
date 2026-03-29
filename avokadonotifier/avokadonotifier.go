// Package avokadonotifier defines notification interfaces and provides factory
// functions that return the correct implementation based on the environment.
package avokadonotifier

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"mime/multipart"
	"net/mail"
	"net/textproto"
	"strings"
	"time"

	"github.com/bilustek/avokado/avokadoerror"
	"github.com/resend/resend-go/v3"
)

// EmailAttachment represents a file to be included with an email message.
type EmailAttachment struct {
	Content     []byte
	Filename    string
	Path        string
	ContentType string
}

// EmailAttachments is a collection of email attachments.
type EmailAttachments []*EmailAttachment

// EmailSenderRequest holds all fields needed to compose and send an email.
type EmailSenderRequest struct {
	From        string
	To          []string
	Subject     string
	Bcc         []string
	Cc          []string
	ReplyTo     string
	HTML        string
	Text        string
	Headers     map[string]string
	Attachments EmailAttachments
}

// EmailSender sends email notifications.
type EmailSender interface {
	Send(ctx context.Context, request *EmailSenderRequest) error
}

// SlackNotifier sends Slack notifications.
type SlackNotifier interface {
	Notify(ctx context.Context, webhookURL, message string) error
}

// EmailSenderRequestToResendRequest converts an EmailSenderRequest into a resend.SendEmailRequest.
func EmailSenderRequestToResendRequest(request *EmailSenderRequest) *resend.SendEmailRequest {
	r := &resend.SendEmailRequest{
		From:    request.From,
		To:      request.To,
		Subject: request.Subject,
		Text:    request.Text,
	}

	if len(request.Bcc) > 0 {
		r.Bcc = request.Bcc
	}
	if len(request.Cc) > 0 {
		r.Cc = request.Cc
	}
	if request.ReplyTo != "" {
		r.ReplyTo = request.ReplyTo
	}
	if request.HTML != "" {
		r.Html = request.HTML
	}
	if len(request.Headers) > 0 {
		r.Headers = request.Headers
	}

	for _, a := range request.Attachments {
		r.Attachments = append(r.Attachments, &resend.Attachment{
			Content:     a.Content,
			Filename:    a.Filename,
			Path:        a.Path,
			ContentType: a.ContentType,
		})
	}

	return r
}

// EmailSenderRequestToMailMessage converts an EmailSenderRequest into a standard net/mail.Message.
func EmailSenderRequestToMailMessage(request *EmailSenderRequest) (*mail.Message, error) {
	header := make(mail.Header)
	header["From"] = []string{request.From}
	header["To"] = []string{strings.Join(request.To, ", ")}
	header["Subject"] = []string{request.Subject}
	header["Date"] = []string{time.Now().Format(time.RFC1123Z)}

	if len(request.Bcc) > 0 {
		header["Bcc"] = []string{strings.Join(request.Bcc, ", ")}
	}
	if len(request.Cc) > 0 {
		header["Cc"] = []string{strings.Join(request.Cc, ", ")}
	}
	if request.ReplyTo != "" {
		header["Reply-To"] = []string{request.ReplyTo}
	}

	for k, v := range request.Headers {
		header[k] = []string{v}
	}

	var buf bytes.Buffer

	contentType := "text/plain; charset=UTF-8"
	content := request.Text
	if request.HTML != "" {
		contentType = "text/html; charset=UTF-8"
		content = request.HTML
	}

	if len(request.Attachments) == 0 {
		header["Content-Type"] = []string{contentType}
		buf.WriteString(content)
	} else {
		mw := multipart.NewWriter(&buf)
		header["Content-Type"] = []string{"multipart/mixed; boundary=" + mw.Boundary()}

		th := make(textproto.MIMEHeader)
		th.Set("Content-Type", contentType)
		p, cpErr := mw.CreatePart(th)
		if cpErr != nil {
			return nil, avokadoerror.New("[EmailSenderRequestToMailMessage multipart.NewWriter CreatePart] err").
				WithErr(cpErr)
		}
		if _, err := p.Write([]byte(content)); err != nil {
			return nil, avokadoerror.New("[EmailSenderRequestToMailMessage] write body part err").WithErr(err)
		}

		for _, a := range request.Attachments {
			ah := make(textproto.MIMEHeader)
			ah.Set("Content-Type", a.ContentType)
			ah.Set("Content-Transfer-Encoding", "base64")
			ah.Set("Content-Disposition", fmt.Sprintf(`attachment; filename=%q`, a.Filename))

			p, cpErr := mw.CreatePart(ah)
			if cpErr != nil {
				return nil, avokadoerror.New("[EmailSenderRequestToMailMessage multipart.NewWriter] Attachments err").
					WithErr(cpErr)
			}
			enc := base64.NewEncoder(base64.StdEncoding, p)
			if _, err := enc.Write(a.Content); err != nil {
				return nil, avokadoerror.New("[EmailSenderRequestToMailMessage] enc write err").WithErr(err)
			}
			_ = enc.Close()
		}
		_ = mw.Close()
	}

	return &mail.Message{
		Header: header,
		Body:   bytes.NewReader(buf.Bytes()),
	}, nil
}
