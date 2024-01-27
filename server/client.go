package main

import (
  // "cmp"
  // "slices"
  "github.com/gofiber/fiber/v2"
  "github.com/daumiller/starkiss/library"
)

// NOTE: temporary, until full multi-user support
var default_user_id string = "8e5624ad-8ce5-4d86-ac5c-1d2ecf120d05"

func startupClientRoutes(server *fiber.App) {
  server.Get("/client/ping",           clientServePing)
  server.Get("/client/categories",     clientServeCategories)
  server.Get("/client/listing/:id",    clientServeListing)
  server.Post("/client/status/:id",    clientSetWatchStatus)
}

func clientServePing(context *fiber.Ctx) error {
  return context.JSON(map[string]string{"message": "pong"})
}

func clientServeCategories(context *fiber.Ctx) error {
  categories, err := library.CategoryList()
  if err != nil { return debug500(context, err) }
  return context.JSON(categories)
}

func clientSetWatchStatus(context *fiber.Ctx) error {
  id := context.Params("id")
  var data struct {
    Started   bool  `json:"started"`
    Timestamp int64 `json:"timestamp"`
  }
  err := context.BodyParser(&data)
  if err != nil { return debug500(context, err) }

  // TODO: run a separate go-routine to debounce these in background, so we're not constantly/unnecessarily updating DB

  md, err := library.MetadataRead(id)
  if err == library.ErrNotFound { return context.SendStatus(404) }
  if err != nil { return debug500(context, err) }

  err = library.UserMetadataSetWatchStatus(default_user_id, md.Id, data.Started, data.Timestamp)
  if err != nil { return debug500(context, err) }

  return context.SendStatus(200)
}

func clientServeListing(context *fiber.Ctx) error {
  id := context.Params("id")
  cat, err := library.CategoryRead(id)
  if (err != nil) && (err != library.ErrNotFound) { return debug500(context, err) }
  if err == nil { return clientServeListing_Category(context, cat) }

  md, err := library.MetadataRead(id)
  if (err != nil) && (err != library.ErrNotFound) { return debug500(context, err) }
  if err == nil { return clientServeListing_Metadata(context, md) }

  return context.SendStatus(404)
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
  Duration  int64  `json:"duration"`
  Started   bool   `json:"started"`
  Timestamp int64  `json:"timestamp"`
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

func clientServeListing_Category(context *fiber.Ctx, cat *library.Category) error {
  children, err := library.UserMetadataViewForParent(default_user_id, cat.Id)
  if err != nil { return debug500(context, err) }

  var listing ClientListing
  listing.Id           = cat.Id
  listing.Name         = cat.Name
  listing.ParentId     = ""
  listing.PosterRatio  = ClientPosterRatio1x1
  listing.EntryCount   = len(children)
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

  // md_ptr := make([]*library.Metadata, len(children))
  // for index := range children { md_ptr[index] = &children[index] }
  // sort_compare := func(a *library.Metadata, b *library.Metadata) int { return cmp.Compare(a.NameSort, b.NameSort) }
  // slices.SortFunc(md_ptr, sort_compare)

  // for index, md := range md_ptr {
  for index, umdv := range children {
    listing.Entries[index].Id        = umdv.MetadataId
    listing.Entries[index].Name      = umdv.NameDisplay
    listing.Entries[index].EntryType = string(umdv.MediaType)
    listing.Entries[index].Duration  = umdv.Duration
    listing.Entries[index].Started   = umdv.Started
    listing.Entries[index].Timestamp = umdv.Timestamp
  }

  return context.JSON(listing)
}

func clientServeListing_Metadata(context *fiber.Ctx, md *library.Metadata) error {
  children, err := library.UserMetadataViewForParent(default_user_id, md.Id)
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

  // why are we sorting here? this list comes from MetadataForParent, which already sorts
  // TODO: verify not needed
  // md_ptr := make([]*library.Metadata, len(children))
  // for index := range children { md_ptr[index] = &children[index] }
  // sort_compare := func(a *library.Metadata, b *library.Metadata) int { return cmp.Compare(a.NameSort, b.NameSort) }
  // slices.SortFunc(md_ptr, sort_compare)

  for index, umdv := range children {
    listing.Entries[index].Id        = umdv.MetadataId
    listing.Entries[index].Name      = umdv.NameDisplay
    listing.Entries[index].EntryType = string(umdv.MediaType)
    listing.Entries[index].Duration  = umdv.Duration
    listing.Entries[index].Started   = umdv.Started
    listing.Entries[index].Timestamp = umdv.Timestamp
  }

  return context.JSON(listing)
}
