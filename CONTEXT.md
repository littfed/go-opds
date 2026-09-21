# OPDS (Open Publication Distribution System)

Core domain model for representing, translating, and serving electronic publication catalogs under OPDS 1.2 (Atom/XML) and OPDS 2.0 (JSON).

## Language

**Feed**:
An Atom-based OPDS 1.2 catalog document composed of metadata, links, and entries.
_Avoid_: Document, stream, channel

**Catalog**:
An OPDS 2.0 JSON document representing a collection of publications, navigation items, and groups.
_Avoid_: Manifest, store, index

**Entry**:
An individual item within an OPDS 1.2 Feed, which can represent either a publication or a navigation link.
_Avoid_: Item, post, record

**Publication**:
A discrete published work (e.g. an eBook) in an OPDS 2.0 Catalog with rich metadata, reading order, and acquisition links.
_Avoid_: Book, article, document

**Navigation**:
A catalog entry or element used to browse hierarchical subsections or categories.
_Avoid_: Category, directory, menu

**Acquisition Link**:
A link relation through which a client can obtain or download a publication.
_Avoid_: Download link, file link

**CatalogProvider**:
An adapter seam that supplies OPDS feeds or catalogs for HTTP request routing.
_Avoid_: FeedHandler, CatalogService
