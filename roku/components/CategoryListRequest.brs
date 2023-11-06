function init()
  m.top.functionName = "requestCategoryList"
end function

function requestCategoryList() as void
  url = CreateObject("roUrlTransfer")
  url.SetUrl(m.global.serverAddress + "/client/categories")
  response = url.GetToString()

  if response.Len() < 0 then
    print "Error retrieving category listing"
    m.top.listing = []
    return
  end if

  json = ParseJson(response)
  if (json = invalid) or (type(json) <> "roArray") then
    print "Error parsing JSON"
    m.top.listing = []
    return
  end if

  m.top.listing = json
  return
end function
