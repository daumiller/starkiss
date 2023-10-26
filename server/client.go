package main

import (
  "cmp"
  "slices"
  "database/sql"
  "github.com/gofiber/fiber/v2"
  "github.com/daumiller/starkiss/database"
)

func startupClientRoutes(server *fiber.App) {
  server.Get("/client/categories", clientServeCategories)
  server.Get("/client/listing/:id", clientServeListing)
}

func clientServeCategories(context *fiber.Ctx) error {
  categories, err := database.CategoryList(DB)
  if err != nil { return debug500(context, err) }
  return context.JSON(categories)
}

func clientServeListing(context *fiber.Ctx) error {
  id := context.Params("id")
  cat, err := database.CategoryRead(DB, id)
  if (err != nil) && (err != database.ErrNotFound) { return debug500(context, err) }
  if err == nil { return clientServeListing_Category(context, &cat) }

  md, err := database.MetadataRead(DB, id)
  if (err != nil) && (err != database.ErrNotFound) { return debug500(context, err) }
  if err == nil { return clientServeListing_Parent(context, &md) }

  return context.SendStatus(404)
}

type ClientPosterRatio string
const (
  ClientPosterRatio1x1  ClientPosterRatio =  "1x1"
  ClientPosterRatio2x3  ClientPosterRatio =  "2x3"
  ClientPosterRatio16x9 ClientPosterRatio = "16x9"
)
type ClientListingEntry struct {
  Id        string `json:"id"`
  Name      string `json:"name"`
  EntryType string `json:"entry_type"`
  Poster    string `json:"poster"`
}
type ClientListing struct {
  Id           string               `json:"id"`
  Name         string               `json:"name"`
  CategoryId   string               `json:"category_id"`
  CategoryType string               `json:"category_type"`
  CategoryName string               `json:"category_name"`
  ParentId     string               `json:"parent_id"`
  ParentName   string               `json:"parent_name"`
  PosterRatio  ClientPosterRatio    `json:"poster_ratio"`
  EntryCount   int                  `json:"entry_count"`
  Entries      []ClientListingEntry `json:"entries"`
}

func clientServeListing_Category(context *fiber.Ctx, cat *database.Category) error {
  metadata, err := database.MetadataWhere(DB, "(parent_id = ?) AND (category_id = ?)", "", cat.Id)
  if err != nil { return debug500(context, err) }

  var listing ClientListing
  listing.Id           = cat.Id
  listing.Name         = cat.Name
  listing.CategoryId   = cat.Id
  listing.CategoryType = string(cat.Type)
  listing.CategoryName = cat.Name
  listing.ParentId     = ""
  listing.ParentName   = ""
  listing.PosterRatio  = ClientPosterRatio1x1
  listing.EntryCount   = len(metadata)
  listing.Entries      = make([]ClientListingEntry, listing.EntryCount)

  switch cat.Type {
    case database.CategoryTypeMovie:  listing.PosterRatio = ClientPosterRatio2x3
    case database.CategoryTypeMusic:  listing.PosterRatio = ClientPosterRatio1x1
    case database.CategoryTypeSeries: listing.PosterRatio = ClientPosterRatio2x3
  }

  md_ptr := make([]*database.Metadata, len(metadata))
  for index := range metadata { md_ptr[index] = &metadata[index] }
  sort_compare := func(a *database.Metadata, b *database.Metadata) int { return cmp.Compare(a.TitleSort, b.TitleSort) }
  slices.SortFunc(md_ptr, sort_compare)

  for index, md := range md_ptr {
    listing.Entries[index].Id        = md.Id
    listing.Entries[index].Name      = md.TitleUser
    listing.Entries[index].EntryType = string(md.Type)
    listing.Entries[index].Poster    = ""
    if md.HasPoster == true {
      listing.Entries[index].Poster = "/poster/" + md.Id + "/small"
    }
  }

  return context.JSON(listing)
}

func clientServeListing_Parent(context *fiber.Ctx, parent *database.Metadata) error {
  var parent_category_name string
  query_row := DB.QueryRow(`SELECT name FROM categories WHERE id = ?;`, parent.CategoryId)
  err := query_row.Scan(&parent_category_name)
  if err == sql.ErrNoRows { return context.SendStatus(404) } // TODO: report that this record is invalid, rather than not found
  if err != nil { return debug500(context, err) }

  parent_parent_name := ""
  if parent.ParentId != "" {
    query_row := DB.QueryRow(`SELECT title_user FROM metadata WHERE id = ?;`, parent.ParentId)
    err := query_row.Scan(&parent_parent_name)
    if err == sql.ErrNoRows { return context.SendStatus(404) } // TODO: report that this record is invalid, rather than not found
    if err != nil { return debug500(context, err) }
  }

  metadata, err := database.MetadataWhere(DB, "(parent_id = ?)", parent.Id)
  if err != nil { return debug500(context, err) }

  var listing ClientListing
  listing.Id           = parent.Id
  listing.Name         = parent.TitleUser
  listing.CategoryId   = parent.CategoryId
  listing.CategoryType = string(parent.CategoryType)
  listing.CategoryName = parent_category_name
  listing.ParentId     = parent.ParentId
  listing.ParentName   = parent_parent_name
  listing.PosterRatio  = ClientPosterRatio1x1
  listing.EntryCount   = len(metadata)
  listing.Entries      = make([]ClientListingEntry, listing.EntryCount)

  scan_series_seasons := false
  switch parent.CategoryType {
    case database.CategoryTypeMovie: listing.PosterRatio = ClientPosterRatio2x3
    case database.CategoryTypeMusic: listing.PosterRatio = ClientPosterRatio1x1
    case database.CategoryTypeSeries:
      listing.PosterRatio = ClientPosterRatio16x9
      scan_series_seasons = true
  }
  // a series listing will display 16x9 if only episodes are present, but display 2x3 if any seasons are present
  if scan_series_seasons == true {
    for _, md := range metadata {
      if md.Type == database.MetadataTypeSeason {
        listing.PosterRatio = ClientPosterRatio2x3
        break
      }
    }
  }

  md_ptr := make([]*database.Metadata, len(metadata))
  for index := range metadata { md_ptr[index] = &metadata[index] }
  sort_compare := func(a *database.Metadata, b *database.Metadata) int { return cmp.Compare(a.TitleSort, b.TitleSort) }
  slices.SortFunc(md_ptr, sort_compare)

  for index, md := range md_ptr {
    listing.Entries[index].Id        = md.Id
    listing.Entries[index].Name      = md.TitleUser
    listing.Entries[index].EntryType = string(md.Type)
    listing.Entries[index].Poster    = ""
    if md.HasPoster == true {
      listing.Entries[index].Poster = "/poster/" + md.Id + "/small"
    }
  }

  return context.JSON(listing)
}
