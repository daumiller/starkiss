function init()
  m.top.functionName = "requestCategoryList"
end function

function requestCategoryList() as void
  url = CreateObject("roUrlTransfer")
  url.SetUrl(m.global.serverAddress + "/client/categories")
  response = url.GetToString()

  if response.Len() < 2 then
    m.top.error = [ "Error retrieving category listing.", "Response was empty." ]
    m.top.listing = []
    return
  end if

  json = ParseJson(response)
  if (json = invalid) or (type(json) <> "roArray") then
    m.top.error = [ "Error retrieving category listing.", "Response was invalid JSON." ]
    m.top.listing = []
    return
  end if

  m.top.error = []
  m.top.listing = json
end function
