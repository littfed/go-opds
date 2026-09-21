# go-opds

[![Go Reference](https://pkg.go.dev/badge/github.com/littfed/go-opds.svg)](https://pkg.go.dev/github.com/littfed/go-opds)
[![Go Report Card](https://goreportcard.com/badge/github.com/littfed/go-opds)](https://goreportcard.com/report/github.com/littfed/go-opds)
[![CI](https://github.com/littfed/go-opds/actions/workflows/ci.yml/badge.svg)](https://github.com/littfed/go-opds/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

[English](#english) | [Русский](#русский)

---

<a name="english"></a>
## English

`go-opds` is a robust, zero-dependency Go library for working with the **OPDS** (Open Publication Distribution System) standard. It supports both **OPDS 1.2** (Atom/XML) and **OPDS 2.0** (JSON), providing invariant-enforcing fluent builders, parsers, symmetric format translators, and a content-negotiating HTTP server.

### Features

- 📚 **OPDS 1.2 Support**: Full Atom/XML feed creation, parsing, and serialization.
- 🚀 **OPDS 2.0 Support**: Full JSON-based catalog creation, parsing, and serialization.
- 🔄 **Symmetric Conversion**: Transparent conversion between OPDS 1.2 Feeds/Entries and OPDS 2.0 Catalogs/Publications.
- 🛠️ **Deep Fluent Builders**: Invariant validation (`Build() (*T, error)`) and ergonomic panic-free shortcuts (`MustBuild() *T`), with auto-detection of MIME types and timestamps.
- 🌐 **Content-Negotiating HTTP Handler**: Built-in `net/http` compatible handler that automatically negotiates OPDS 1.2 vs 2.0 based on the client's `Accept` header and translates formats on the fly.
- ⚡ **Zero External Dependencies**: Implemented strictly with the Go standard library.

### Installation

```bash
go get github.com/littfed/go-opds
```

### Quick Start

#### 1. Building an OPDS 1.2 Feed (Atom / XML)

```go
package main

import (
	"fmt"
	"time"

	"github.com/littfed/go-opds/opds1"
)

func main() {
	feed := opds1.NewFeedBuilder(
		"urn:uuid:my-catalog",
		"My E-Book Catalog",
		time.Now().UTC().Format(time.RFC3339),
	).
		Author("Catalog Admin", "https://example.com").
		AddLink(opds1.NewLinkBuilder().
			Rel(opds1.RelSelf).
			Href("/opds").
			Type(opds1.TypeNavigation).
			Build()).
		AddEntry(opds1.NewEntryBuilder(
			"urn:isbn:978-0-123456-47-2",
			"Sample Book Title",
			time.Now().UTC().Format(time.RFC3339),
		).
			Author("Author Name", "").
			Summary("A description of the book.").
			AddLink(opds1.NewLinkBuilder().
				Rel(opds1.RelAcquisition).
				Href("/files/book.epub"). // Auto-detects application/epub+zip!
				Build()).
			MustBuild()).
		MustBuild()

	xmlData, err := feed.ToXML()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(xmlData))
}
```

#### 2. Building an OPDS 2.0 Catalog (JSON)

```go
package main

import (
	"fmt"

	"github.com/littfed/go-opds/opds2"
)

func main() {
	catalog := opds2.NewCatalogBuilderWithID(
		"urn:uuid:my-catalog",
		"My Modern OPDS 2.0 Catalog",
	).
		AddLink(opds2.NewLinkBuilder().
			Rel(opds2.RelSelf).
			Href("/opds/v2").
			Type(opds2.TypeNavigation).
			Build()).
		AddPublication(opds2.NewPublicationBuilderWithID(
			"urn:isbn:978-0-123456-47-2",
			"Sample Book Title",
		).
			Author("Author Name", "").
			Description("A description of the book.").
			AddAcquisitionLink("/files/book.epub", ""). // Auto-detects media type
			MustBuild()).
		MustBuild()

	jsonData, err := catalog.ToJSON()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(jsonData))
}
```

#### 3. Converting Between OPDS 1.2 and OPDS 2.0

```go
import "github.com/littfed/go-opds/convert"

// Full catalog conversion:
opds2Catalog := convert.Catalog1To2(opds1Feed)
opds1Feed    := convert.Catalog2To1(opds2Catalog)

// Granular entry / publication conversion:
opds2Pub  := convert.EntryToPublication(opds1Entry)
opds1Item := convert.PublicationToEntry(opds2Pub)
```

#### 4. Serving via HTTP with Content Negotiation

The handler automatically inspects the client's `Accept` header and translates the feed into the requested OPDS format on the fly:

```go
package main

import (
	"net/http"

	"github.com/littfed/go-opds/handler"
	"github.com/littfed/go-opds/opds1"
)

func main() {
	h := handler.New(handler.Config{
		BaseURL:        "http://localhost:8080",
		DefaultVersion: "1.2", // Fallback when Accept header is missing
		FeedFunc: func(path string) (any, error) {
			// Return either *opds1.Feed or *opds2.Catalog - handler auto-translates!
			return opds1.NewFeedBuilder("urn:root", "My Library", "").MustBuild(), nil
		},
	})

	http.ListenAndServe(":8080", h)
}
```

### Examples

The repository includes ready-to-run examples in the [`examples/`](examples) directory:

- [`examples/basic`](examples/basic): Demonstrates creating, converting, and serving OPDS 1.2 and 2.0 feeds over HTTP.
- [`examples/client`](examples/client): A complete OPDS client with feed navigation and book downloading.

Run the basic server example:
```bash
go run ./examples/basic
```

### Testing

Run all unit tests:
```bash
go test -v ./...
```

---

<a name="русский"></a>
## Русский

`go-opds` — это надежная библиотека на языке Go без внешних зависимостей для работы со стандартами **OPDS** (Open Publication Distribution System). Поддерживаются обе версии: **OPDS 1.2** (Atom/XML) и **OPDS 2.0** (JSON).

### Возможности

- 📚 **Поддержка OPDS 1.2**: Создание, парсинг и валидация Atom/XML каталогов.
- 🚀 **Поддержка OPDS 2.0**: Создание, парсинг и сериализация JSON-каталогов.
- 🔄 **Симметричная конвертация**: Прозрачное преобразование фидов и отдельных книг между версиями 1.2 и 2.0.
- 🛠️ **Глубокие Fluent Builders**: Валидация инвариантов протокола (`Build() (*T, error)`) и удобные шорткаты (`MustBuild() *T`), автоматический вывод MIME-типов и RFC3339 дат.
- 🌐 **HTTP Handler c Content Negotiation**: Обработчик с согласованием содержимого (заголовок `Accept` читалки) и автоматической трансляцией форматов на лету.
- ⚡ **Без внешних зависимостей**: Написано только на стандартной библиотеке Go.

### Установка

```bash
go get github.com/littfed/go-opds
```

### Примеры использования

#### 1. Создание каталога OPDS 1.2 (Atom / XML)

```go
package main

import (
	"fmt"
	"time"

	"github.com/littfed/go-opds/opds1"
)

func main() {
	feed := opds1.NewFeedBuilder(
		"urn:uuid:my-catalog",
		"Мой каталог книг",
		time.Now().UTC().Format(time.RFC3339),
	).
		Author("Администратор", "https://example.com").
		AddLink(opds1.NewLinkBuilder().
			Rel(opds1.RelSelf).
			Href("/opds").
			Type(opds1.TypeNavigation).
			Build()).
		AddEntry(opds1.NewEntryBuilder(
			"urn:isbn:978-0-123456-47-2",
			"Название книги",
			time.Now().UTC().Format(time.RFC3339),
		).
			Author("Автор", "").
			Summary("Краткое описание книги.").
			AddLink(opds1.NewLinkBuilder().
				Rel(opds1.RelAcquisition).
				Href("/files/book.epub"). // MIME-тип определится автоматически!
				Build()).
			MustBuild()).
		MustBuild()

	xmlData, err := feed.ToXML()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(xmlData))
}
```

#### 2. Создание каталога OPDS 2.0 (JSON)

```go
package main

import (
	"fmt"

	"github.com/littfed/go-opds/opds2"
)

func main() {
	catalog := opds2.NewCatalogBuilderWithID(
		"urn:uuid:my-catalog",
		"Современный каталог OPDS 2.0",
	).
		AddLink(opds2.NewLinkBuilder().
			Rel(opds2.RelSelf).
			Href("/opds/v2").
			Type(opds2.TypeNavigation).
			Build()).
		AddPublication(opds2.NewPublicationBuilderWithID(
			"urn:isbn:978-0-123456-47-2",
			"Название книги",
		).
			Author("Автор", "").
			Description("Описание книги.").
			AddAcquisitionLink("/files/book.epub", ""). // Авто-определение MIME
			MustBuild()).
		MustBuild()

	jsonData, err := catalog.ToJSON()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(jsonData))
}
```

#### 3. Конвертация между OPDS 1.2 и OPDS 2.0

```go
import "github.com/littfed/go-opds/convert"

// Конвертация каталогов целиком:
opds2Catalog := convert.Catalog1To2(opds1Feed)
opds1Feed    := convert.Catalog2To1(opds2Catalog)

// Гранулярная конвертация отдельных записей / книг:
opds2Pub  := convert.EntryToPublication(opds1Entry)
opds1Item := convert.PublicationToEntry(opds2Pub)
```

#### 4. Раздача по HTTP с автоматическим Content Negotiation

```go
package main

import (
	"net/http"

	"github.com/littfed/go-opds/handler"
	"github.com/littfed/go-opds/opds1"
)

func main() {
	h := handler.New(handler.Config{
		BaseURL:        "http://localhost:8080",
		DefaultVersion: "1.2",
		FeedFunc: func(path string) (any, error) {
			// Можно вернуть *opds1.Feed или *opds2.Catalog — сервер сам транслирует под читателя!
			return opds1.NewFeedBuilder("urn:root", "Библиотека", "").MustBuild(), nil
		},
	})

	http.ListenAndServe(":8080", h)
}
```

### Запуск примеров и тестов

Примеры доступны в каталоге [`examples/`](examples):
- [`examples/basic`](examples/basic): Сервер с поддержкой OPDS 1.2 и 2.0 и конвертацией на лету.
- [`examples/client`](examples/client): Полноценный консольный OPDS-клиент.

Запуск сервера примеров:
```bash
go run ./examples/basic
```

Запуск тестов:
```bash
go test -v ./...
```

---

## License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.
