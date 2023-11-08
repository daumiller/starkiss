function init()
  m.top.functionName = "requestMediaList"
end function

function requestMediaList() as void
  contentNode = CreateObject("roSGNode", "ContentNode")
  contentNode.SetFields({ role: "Content" })

  url = CreateObject("roUrlTransfer")
  url.SetUrl(m.global.serverAddress + "/client/listing/" + m.top.id)
  response = url.GetToString()

  if response.Len() < 0 then
    print "Error retrieving media listing"
    m.top.content     = contentNode
    m.top.parentId    = ""
    m.top.title       = "(error)"
    m.top.posterRatio = "2x3"
    m.top.entryCount  = 0
    m.top.entries     = []
    return
  end if

  json = ParseJson(response)
  if (json = invalid) or (type(json) <> "roAssociativeArray") then
    print "Error parsing JSON"
    m.top.content     = contentNode
    m.top.parentId    = ""
    m.top.title       = "(error)"
    m.top.posterRatio = "2x3"
    m.top.entryCount  = 0
    m.top.entries     = []
    return
  end if

  m.top.parentId    = json.parent_id
  m.top.title       = json.name
  m.top.posterRatio = json.poster_ratio
  m.top.listingType = json.listing_type
  m.top.entryCount  = json.entry_count
  m.top.entries     = json.entries

  for each entry in json.entries
    node = CreateObject("roSGNode", "ContentNode")

    if (m.top.listingType = "episodes") or (m.top.listingType = "songs") then
      node.SetFields({
        shortDescriptionLine1: entry.entry_type,
        shortDescriptionLine2: entry.id,
        title: entry.name,
      })
    else
      node.SetFields({
        shortDescriptionLine1: entry.name,
        shortDescriptionLine2: entry.id,
        title: entry.entry_type,
        hdPosterUrl: m.global.serverAddress + "/poster/" + entry.id + "/small",
      })
    end if
    contentNode.AppendChild(node)
    ' print "Added " + entry.name
    ' print "Type " + entry.entry_type
  end for

  m.top.content = contentNode
  return
end function
