package main

import (
  "cmp"
  "slices"
  "github.com/gofiber/fiber/v2"
  "github.com/daumiller/starkiss/library"
)

func startupClientRoutes(server *fiber.App) {
  server.Get("/client/categories", clientServeCategories)
  server.Get("/client/listing/:id", clientServeListing)
}

func clientServeCategories(context *fiber.Ctx) error {
  categories, err := library.CategoryList()
  if err != nil { return debug500(context, err) }
  return context.JSON(categories)
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
  EntryCount   int                  `json:"entry_count"`
  Entries      []ClientListingEntry `json:"entries"`
}

func clientServeListing_Category(context *fiber.Ctx, cat *library.Category) error {
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

  md_ptr := make([]*library.Metadata, len(metadata))
  for index := range metadata { md_ptr[index] = &metadata[index] }
  sort_compare := func(a *library.Metadata, b *library.Metadata) int { return cmp.Compare(a.NameSort, b.NameSort) }
  slices.SortFunc(md_ptr, sort_compare)

  for index, md := range md_ptr {
    listing.Entries[index].Id        = md.Id
    listing.Entries[index].Name      = md.NameDisplay
    listing.Entries[index].EntryType = string(md.MediaType)
  }

  return context.JSON(listing)
}

func clientServeListing_Metadata(context *fiber.Ctx, md *library.Metadata) error {
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

  md_ptr := make([]*library.Metadata, len(children))
  for index := range children { md_ptr[index] = &children[index] }
  sort_compare := func(a *library.Metadata, b *library.Metadata) int { return cmp.Compare(a.NameSort, b.NameSort) }
  slices.SortFunc(md_ptr, sort_compare)

  for index, md := range md_ptr {
    listing.Entries[index].Id        = md.Id
    listing.Entries[index].Name      = md.NameDisplay
    listing.Entries[index].EntryType = string(md.MediaType)
  }

  return context.JSON(listing)
}
