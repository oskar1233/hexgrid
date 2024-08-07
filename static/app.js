const tilesContainer = document.getElementById("tiles-container")

const websocket = new WebSocket("wss://squaregrid.oskar1233.dev/ws")

let color = null
let width = 0
let height = 0

websocket.addEventListener("message", e => {
  e.data.arrayBuffer().then(buffer => {
    const dataView = new DataView(buffer)
    const command = dataView.getInt8(0)

    if (command === 1) {
      const width = dataView.getUint16(1)
      const height = dataView.getUint16(3)

      const colorR = dataView.getUint8(5)
      const colorG = dataView.getUint8(6)
      const colorB = dataView.getUint8(7)

      setTilesWH(width, height)
      setColor(colorR, colorG, colorB)

      const commandByte = new Uint8Array(1)
      commandByte[0] = 3

      const blob = new Blob([commandByte])
      websocket.send(blob)
    } else if (command == 3) {
      for (let i = 0; i < width*height; ++i) {
        const colorR = dataView.getUint8((i*3)+1)
        const colorG = dataView.getUint8((i*3)+2)
        const colorB = dataView.getUint8((i*3)+3)

        tilesContainer.children[i].style.background = colorToHex(colorR, colorG, colorB)
      }
    } else {
      console.error("Invalid init message (invalid first byte)", command)
    } 
  })
})

function setTilesWH(newWidth, newHeight) {
  width = newWidth
  height = newHeight

  for(let i = 0; i < width*height; ++i) {
    const elem = document.createElement("div")
    tilesContainer.append(elem)
  }
}

function colorToHex(r, g, b) {
  return "#" + r.toString(16) + g.toString(16) + b.toString(16)
}

function setColor(r, g, b) {
  color = colorToHex(r, g, b)

  document.getElementById("your-color").style.background = color
}

tilesContainer.addEventListener("click", e => {
  if(websocket.readyState !== WebSocket.OPEN) {
    location.reload()
  }

  if(e.target !== tilesContainer) {
    const commandByte = new Uint8Array(1)
    commandByte[0] = 2

    const offsetBytes = new Uint32Array(1)
    offsetBytes[0] = Array.prototype.indexOf.call(e.target.parentNode.children, e.target)

    const blob = new Blob([commandByte, offsetBytes])
    websocket.send(blob)

    if(color) {
      e.target.style.background = color
    }
  }
})
