let ws = new WebSocket("ws://localhost:5309/ws")

ws.onopen = (_) => { }

ws.onmessage = (event) => {
	// console.log(`[message] data received from server: ${event.data}`);
	const data = JSON.parse(event.data)
	switch (data.command) {
		case "clear":
			document.getElementById("result").innerHTML = ""
			break
		case "replaced":
			document.getElementById("result").innerHTML = data.payload.script
			break
		case "exec":
			document.getElementById("result").innerHTML += data.payload.fragment
			break
		default:
			console.log(`[message] unrecognized command: ${data.command}`)
	}
}

ws.onclose = (event) => {
	if (event.wasClean) {
		console.log(`[close] connection closed cleanly, code=${event.code}, reason=${event.reason}`)
	} else {
		console.log(`[close] connection died, event=${JSON.stringify(event)}`)
	}
}

ws.onerror = (error) => {
	console.log(`[error] error=${JSON.stringify(error)}`)
}

const viewRequest = (path, title, scope, environment) => {
	ws.send(JSON.stringify({
		command: "view",
		payload: {
			path: path,
			title: title,
			scope: scope,
			environment: environment,
		}
	}))
}

const execRequest = (path, title, scope, environment) => {
	ws.send(JSON.stringify({
		command: "exec",
		payload: {
			path: path,
			title: title,
			scope: scope,
			environment: environment,
		}
	}))
}
