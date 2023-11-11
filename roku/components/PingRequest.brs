function init()
  m.top.functionName = "requestPing"
end function

function requestPing() as void
  url = CreateObject("roUrlTransfer")
  url.SetUrl(m.top.address + "/client/ping")
  response = url.GetToString()

  if response.Len() < 2 then
    m.top.success = false
    return
  end if

  json = ParseJson(response)
  if (json = invalid) or (type(json) <> "roAssociativeArray") then
    m.top.success = false
    return
  end if

  if json.DoesExist("message") = false then
    m.top.success = false
    return
  end if

  if json.message <> "pong" then
    m.top.success = false
    return
  end if

  m.top.success = true
end function
