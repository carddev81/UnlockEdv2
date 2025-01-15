package models

import (
	"encoding/xml"
	"fmt"
	"strings"
)

type Library struct {
	DatabaseFields
	OpenContentProviderID uint    `gorm:"not null" json:"open_content_provider_id"`
	ExternalID            *string `json:"external_id"`
	Title                 string  `gorm:"size:255;not null" json:"title"`
	Language              *string `gorm:"size:255" json:"language"`
	Description           *string `json:"description"`
	Url                   string  `gorm:"not null" json:"url"`
	ThumbnailUrl          *string `json:"thumbnail_url"`
	VisibilityStatus      bool    `gorm:"default:false;not null" json:"visibility_status"`

	OpenContentProvider *OpenContentProvider  `gorm:"foreignKey:OpenContentProviderID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"open_content_provider"`
	Favorites           []OpenContentFavorite `gorm:"-" json:"favorites"`
}

func (Library) TableName() string { return "libraries" }

func (lib *Library) IntoProxyPO() *LibraryProxyPO {
	proxyParams := LibraryProxyPO{
		ID:                    lib.ID,
		Path:                  lib.Url,
		BaseUrl:               lib.OpenContentProvider.Url,
		OpenContentProviderID: lib.OpenContentProvider.ID,
		VisibilityStatus:      lib.VisibilityStatus,
	}
	return &proxyParams
}

type LibraryProxyPO struct {
	ID                    uint
	OpenContentProviderID uint
	Path                  string
	BaseUrl               string
	VisibilityStatus      bool
}

type KiwixChannel struct {
	Title        string      `json:"title"`
	ThumbnailUrl string      `json:"thumbnail_url"`
	Link         string      `json:"link"`
	Description  string      `json:"description"`
	TotalResults string      `json:"total_results"`
	StartIndex   string      `json:"start_index"`
	ItemsPerPage string      `json:"items_per_page"`
	Items        []KiwixItem `json:"items"`
}

type KiwixItem struct {
	Title        string `json:"title"`
	Link         string `json:"link"`
	Description  string `json:"description"`
	Book         string `json:"book"`
	ThumbnailUrl string `json:"thumbnail_url"`
	WordCount    string `json:"word_count"`
}

func (rss *RSS) IntoKiwixChannel(library *Library) *KiwixChannel {
	channel := &KiwixChannel{
		Title:        rss.Channel.Title,
		Link:         rss.Channel.Link,
		ThumbnailUrl: *library.ThumbnailUrl,
		Description:  rss.Channel.Description,
		TotalResults: rss.Channel.TotalResults,
		StartIndex:   rss.Channel.StartIndex,
		ItemsPerPage: rss.Channel.ItemsPerPage,
		Items:        []KiwixItem{},
	}

	for _, item := range rss.Channel.Items {
		// Prepare the bolded description
		///descriptionWithBold := formatBoldDescription(item.Description.RawText, item.Description.Bold)
		descriptionWithBold := item.Description.RawText
		kiwixItem := KiwixItem{
			Title:        item.Title,
			Link:         fmt.Sprintf("/api/proxy/libraries/%d%s", library.ID, item.Link),
			ThumbnailUrl: *library.ThumbnailUrl,
			Description:  descriptionWithBold,
			Book:         item.Book.Title,
			WordCount:    item.WordCount,
		}
		channel.Items = append(channel.Items, kiwixItem)
	}
	return channel
}

// XML Parsing START for kiwix here...
type RSS struct {
	XMLName    xml.Name `xml:"rss"`
	Version    string   `xml:"version,attr"`
	OpenSearch string   `xml:"xmlns:opensearch,attr"`
	Atom       string   `xml:"xmlns:atom,attr"`
	Channel    Channel  `xml:"channel"`
}



type Channel struct {
	Title        string          `xml:"title"`
	Link         string          `xml:"link"`
	Description  string          `xml:"description"`
	TotalResults string          `xml:"http://a9.com/-/spec/opensearch/1.1/ totalResults"`
	StartIndex   string          `xml:"http://a9.com/-/spec/opensearch/1.1/ startIndex"`
	ItemsPerPage string          `xml:"http://a9.com/-/spec/opensearch/1.1/ itemsPerPage"`
	SearchLink   AtomLink        `xml:"atom:link"`
	Query        OpenSearchQuery `xml:"opensearch:Query"`
	Items        []Item          `xml:"item"`
}

type AtomLink struct {
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
	Href string `xml:"href,attr"`
}

type OpenSearchQuery struct {
	Role        string `xml:"role,attr"`
	SearchTerms string `xml:"searchTerms,attr"`
	StartIndex  string `xml:"startIndex,attr"`
	Count       string `xml:"count,attr"`
}

type Item struct {
	Title       string      `xml:"title"`
	Link        string      `xml:"link"`
	Description Description `xml:"description"`
	Book        Book        `xml:"book"`
	WordCount   string      `xml:"wordCount"`
}

type Description struct {
	RawText string   `xml:"-"`
	Bold    []string `xml:"-"`
}

type Book struct {
	Title string `xml:"title"`
}
//for retaining bolded words from search had to implement this function to decode the indiviual tag tokens
func (d *Description) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	var rawContent strings.Builder
	var boldWords []string

	for {
		tok, err := dec.Token()
		if err != nil {
			return err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "b" {
				var boldText string
				dec.DecodeElement(&boldText, &t)
				boldWords = append(boldWords, boldText)
				rawContent.WriteString("<b>" + boldText + "</b>")
			} else {
				rawContent.WriteString("<" + t.Name.Local + ">")
			}
		case xml.EndElement:
			if t.Name.Local == start.Name.Local {
				d.RawText = rawContent.String()
				d.Bold = boldWords
				return nil
			}
			rawContent.WriteString("</" + t.Name.Local + ">")
		case xml.CharData:
			rawContent.WriteString(string(t))
		}
	}
}
