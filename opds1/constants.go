package opds1

import "github.com/littfed/go-opds/internal/common"

// MIME types for OPDS catalog feed documents.
const (
	TypeNavigation  = `application/atom+xml;profile=opds-catalog;kind=navigation`
	TypeAcquisition = `application/atom+xml;profile=opds-catalog;kind=acquisition`
	TypeEntry       = `application/atom+xml;type=entry;profile=opds-catalog`
)

// MIME types for publication formats.
const (
	MIMEEPUB = common.MIMEEPUB
	MIMEPDF  = common.MIMEPDF
	MIMEMOBI = common.MIMEMOBI
	MIMEAZW3 = common.MIMEAZW3
	MIMECBZ  = common.MIMECBZ
	MIMEFB2  = common.MIMEFB2
	MIMEHTML = common.MIMEHTML
	MIMEPNG  = common.MIMEPNG
	MIMEJPEG = common.MIMEJPEG
	MIMEGIF  = common.MIMEGIF
)

// Standard link relations.
const (
	RelStart             = "start"
	RelSelf              = "self"
	RelNext              = "next"
	RelPrevious          = "prev"
	RelUp                = "up"
	RelRelated           = "related"
	RelSubsection        = "subsection"
	RelSearch            = "search"
	RelAlternate         = "alternate"
	RelAcquisition       = "http://opds-spec.org/acquisition"
	RelAcquisitionBuy    = "http://opds-spec.org/acquisition/buy"
	RelAcquisitionBorrow = "http://opds-spec.org/acquisition/borrow"
	RelAcquisitionOpenAccess = "http://opds-spec.org/acquisition/open-access"
	RelAcquisitionSample = "http://opds-spec.org/acquisition/sample"
	RelAcquisitionSubscribe = "http://opds-spec.org/acquisition/subscribe"
	RelImage             = "http://opds-spec.org/image"
	RelImageThumbnail    = "http://opds-spec.org/image/thumbnail"
	RelSortNew           = "http://opds-spec.org/sort/new"
	RelSortPopular       = "http://opds-spec.org/sort/popular"
	RelFeatured          = "http://opds-spec.org/featured"
	RelRecommended       = "http://opds-spec.org/recommended"
	RelCrawlable         = "http://opds-spec.org/crawlable"
	RelShelf             = "http://opds-spec.org/shelf"
	RelSubscriptions     = "http://opds-spec.org/subscriptions"
	RelFacet             = "http://opds-spec.org/facet"
	RelCollection        = "collection"
)
