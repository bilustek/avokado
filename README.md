![Version](https://img.shields.io/badge/version-0.0.0-orange.svg)
![Go](https://img.shields.io/github/go-mod/go-version/bilustek/avokado)
[![Documentation](https://godoc.org/github.com/bilustek/avokado?status.svg)](https://pkg.go.dev/github.com/bilustek/avokado)
[![Go tests](https://github.com/bilustek/avokado/actions/workflows/go-test.yml/badge.svg)](https://github.com/bilustek/avokado/actions/workflows/go-test.yml)
[![Go lint](https://github.com/bilustek/avokado/actions/workflows/go-lint.yml/badge.svg)](https://github.com/bilustek/avokado/actions/workflows/go-lint.yml)
[![codecov](https://codecov.io/github/bilustek/avokado/graph/badge.svg?token=4SKJU2QH42)](https://codecov.io/github/bilustek/avokado)


# avokado

Quick rest api server genarator ([Django](https://www.djangoproject.com/) inspired!) for nerds! Batteries included:

- [Fiber V3](https://github.com/gofiber/fiber)
- [Sentry](https://github.com/getsentry/sentry-go)
- [Validator](https://github.com/go-playground/validator)
- [PostgreSQL](https://www.postgresql.org/)
- [Gorm](https://gorm.io/)
- [Golang Migrate](https://github.com/golang-migrate/migrate/)
- [Scalar](https://scalar.com/)

## Features

- Django style migrations, extend your models from `BaseModel`
- Django style User, Group, Permission approach.
- Django style User register, confirm, change password approach
- Django style admin endpoints
- Sign-in With Google (provider based authentication)
- Sign-in With Password (password based authentication)
- Slack, Email or console backends based on environment *
- OpenAPI 3.0 specification and Scalar UI
- Ruby on Rails style **strong params** implementation

## Built-in Models

- `auth_user`
- `auth_group`
- `auth_permission`
- `auth_user_groups`
- `auth_user_permissions`
- `auth_group_permissions`
- `refresh_token`

@wip

---

## Requirements

@wip

---

## Installation

@wip

---

## Usage

### Management Commands

- `showmigrations`
- `makemigrations`
- `migrate`

```bash
go run ./cmd/management/showmigrations/
go run ./cmd/management/showmigrations/ -h

go run ./cmd/management/makemigrations/ -h
go run ./cmd/management/makemigrations/ --name "<name-of-your-migration>"

go run ./cmd/management/migrate/ -h
go run ./cmd/management/migrate/
```

---

## Contributor(s)

- [Uğur Özyılmazel](https://github.com/vigo) - Creator, maintainer

---

## Contribute

All PR’s are welcome!

1. `fork` (https://github.com/bilustek/avokado/fork)
1. Create your `branch` (`git checkout -b my-feature`)
1. `commit` yours (`git commit -am 'add some functionality'`)
1. `push` your `branch` (`git push origin my-feature`)
1. Than create a new **Pull Request**!

---

## License

This project is licensed under MIT

---

This project is intended to be a safe, welcoming space for collaboration, and
contributors are expected to adhere to the [code of conduct][coc].

[coc]: https://github.com/bilustek/avokado/blob/main/CODE_OF_CONDUCT.md
