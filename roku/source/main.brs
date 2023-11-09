sub Main()
  screen = CreateObject("roSGScreen")
  m.port = CreateObject("roMessagePort")
  screen.setMessagePort(m.port)

  scene = screen.CreateScene("Main")
  globals = screen.getGlobalNode()
  globals.addFields({
    "serverAddress": "http://192.168.0.250:4331",
    "autoPlay": true,
  })

  screen.show()
  scene.setFocus(true)

  print "Starkiss entering main loop..."
  while(true)
    msg = wait(0, m.port)
    msgType = type(msg)
    if msgType = "roSGScreenEvent" then
      if msg.isScreenClosed() then return
    end if
  end while
  print "Starkiss exiting..."
end sub
