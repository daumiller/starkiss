// filepath: /Users/darcy/dev/starkiss/server/client.go
package main

import (
  "cmp"
  "slices"

  "github.com/labstack/echo/v4"
  "github.com/daumiller/starkiss/library"
)

func startupClientRoutes(server *echo.Echo) {
  server.GET("/client/ping",        clientServePing)
  server.GET("/client/categories",  clientServeCategories)
  server.GET("/client/listing/:id", clientServeListing)
}

func clientServePing(context echo.Context) error {
  return context.JSON(200, map[string]string{"message": "pong"})
}

func clientServeCategories(context echo.Context) error {
  categories, err := library.CategoryList()
  if err != nil { return debug500(context, err) }
  return context.JSON(200, categories)
}

func clientServeListing(context echo.Context) error {
  id := context.Param("id")
  cat, err := library.CategoryRead(id)
  if (err != nil) && (err != library.ErrNotFound) { return debug500(context, err) }
  if err == nil { return clientServeListing_Category(context, cat) }

  md, err := library.MetadataRead(id)
  if (err != nil) && (err != library.ErrNotFound) { return debug500(context, err) }
  if err == nil { return clientServeListing_Metadata(context, md) }

  return context.NoContent(404)
}

type ClientPosterRatio string
const (
  ClientPosterRatio1x1  ClientPosterRatio =  "1x1"
  ClientPosterRatio2x3  ClientPosterRatio =  "2x3"
)
type ClientListingType string
const (
  ClientListingTypeMovies   ClientListingType = "movies"
  ClientListingTypeSeries   ClientListingType = "series"
  ClientListingTypeSeasons  ClientListingType = "seasons"
  ClientListingTypeEpisodes ClientListingType = "episodes"
  ClientListingTypeArtists  ClientListingType = "artists"
  ClientListingTypeAlbums   ClientListingType = "albums"
  ClientListingTypeSongs    ClientListingType = "songs"
  ClientListingTypeInvalid  ClientListingType = "invalid"
)
type ClientListingEntry struct {
  Id        string `json:"id"`
  Name      string `json:"name"`
  EntryType string `json:"entry_type"`
}
type ClientListing struct {
  Id           string               `json:"id"`
  Name         string               `json:"name"`
  ParentId     string               `json:"parent_id"`
  PosterRatio  ClientPosterRatio    `json:"poster_ratio"`
  ListingType  ClientListingType    `json:"listing_type"`
  EntryCount   int                  `json:"entry_count"`
  Entries      []ClientListingEntry `json:"entries"`
}

func clientServeListing_Category(context echo.Context, cat *library.Category) error {
  metadata, err := library.MetadataForParent(cat.Id)
  if err != nil { return debug500(context, err) }

  var listing ClientListing
  listing.Id           = cat.Id
  listing.Name         = cat.Name
  listing.ParentId     = ""
  listing.PosterRatio  = ClientPosterRatio1x1
  listing.EntryCount   = len(metadata)
  listing.Entries      = make([]ClientListingEntry, listing.EntryCount)

  switch cat.MediaType {
    case library.CategoryMediaTypeSeries : fallthrough
    case library.CategoryMediaTypeMovie  : listing.PosterRatio = ClientPosterRatio2x3
  }

  switch cat.MediaType {
    case library.CategoryMediaTypeMovie  : listing.ListingType = ClientListingTypeMovies
    case library.CategoryMediaTypeSeries : listing.ListingType = ClientListingTypeSeries
    case library.CategoryMediaTypeMusic  : listing.ListingType = ClientListingTypeArtists
  }

  md_ptr := make([]*library.Metadata, len(metadata))
  for index := range metadata { md_ptr[index] = &metadata[index] }
  sort_compare := func(a *library.Metadata, b *library.Metadata) int { return cmp.Compare(a.NameSort, b.NameSort) }
  slices.SortFunc(md_ptr, sort_compare)

  for index, md := range md_ptr {
    listing.Entries[index].Id        = md.Id
    listing.Entries[index].Name      = md.NameDisplay
    listing.Entries[index].EntryType = string(md.MediaType)
  }

  return context.JSON(200, listing)
}

func clientServeListing_Metadata(context echo.Context, md *library.Metadata) error {
  children, err := library.MetadataForParent(md.Id)
  if err != nil { return debug500(context, err) }

  var listing ClientListing
  listing.Id           = md.Id
  listing.Name         = md.NameDisplay
  listing.ParentId     = md.ParentId
  listing.PosterRatio  = ClientPosterRatio1x1
  listing.EntryCount   = len(children)
  listing.Entries      = make([]ClientListingEntry, listing.EntryCount)

  switch md.MediaType {
    case library.MetadataMediaTypeSeason: fallthrough
    case library.MetadataMediaTypeSeries: listing.PosterRatio = ClientPosterRatio2x3
  }

  switch md.MediaType {
    case library.MetadataMediaTypeFileVideo : listing.ListingType = ClientListingTypeInvalid
    case library.MetadataMediaTypeFileAudio : listing.ListingType = ClientListingTypeInvalid
    case library.MetadataMediaTypeSeries    : listing.ListingType = ClientListingTypeSeasons
    case library.MetadataMediaTypeSeason    : listing.ListingType = ClientListingTypeEpisodes
    case library.MetadataMediaTypeArtist    : listing.ListingType = ClientListingTypeAlbums
    case library.MetadataMediaTypeAlbum     : listing.ListingType = ClientListingTypeSongs
  }

  // A "Series" metadata may be a parent to Seasons, or to Episodes (but not both (yet)).
  if listing.ListingType == ClientListingTypeSeasons {
    if len(children) > 0 {
      if children[0].MediaType == library.MetadataMediaTypeFileVideo {
        listing.ListingType = ClientListingTypeEpisodes
      }
    }
  }

  // An Artist metadata may be parent to Albums, or to Songs (but not both (yet)).
  if listing.ListingType == ClientListingTypeAlbums {
    if len(children) > 0 {
      if children[0].MediaType == library.MetadataMediaTypeFileAudio {
        listing.ListingType = ClientListingTypeSongs
      }
    }
  }

  md_ptr := make([]*library.Metadata, len(children))
  for index := range children { md_ptr[index] = &children[index] }
  sort_compare := func(a *library.Metadata, b *library.Metadata) int { return cmp.Compare(a.NameSort, b.NameSort) }
  slices.SortFunc(md_ptr, sort_compare)

  for index, md := range md_ptr {
    listing.Entries[index].Id        = md.Id
    listing.Entries[index].Name      = md.NameDisplay
    listing.Entries[index].EntryType = string(md.MediaType)
  }

  return context.JSON(200, listing)
}
