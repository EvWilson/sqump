import { StreamLanguage } from "@codemirror/language"
import { EditorView } from "@codemirror/view"
import { lua } from "@codemirror/legacy-modes/mode/lua"
import { defaultSetup } from "./shared.mjs"

const luaSource = document.querySelector("#lua-editarea")
const luaTarget = document.querySelector("#lua-editor")

const luaSync = EditorView.updateListener.of((v) => {
  console.log("sync change:", JSON.stringify(v))
  if (v.docChanged) {
    // Document changed
  }
})

// Messing with state seemed to be trouble
// Mostly took this config from https://codemirror.net/examples/config/, near the end
let luaView = new EditorView({
  doc: luaSource.value,
  extensions: [
    defaultSetup,
    StreamLanguage.define(lua),
    luaSync
  ],
  parent: luaTarget,
})

// Little trick from https://discuss.codemirror.net/t/codemirror-6-and-textareas/2731/3
// const syncLuaEditor = () => {
//   luaSource.value = luaView.state.sliceDoc()
//   console.log(luaSource.value)
// }

// document.querySelector("#lua-submit").addEventListener("click", syncLuaEditor)
// document.querySelector("#lua-exec").addEventListener("click", syncLuaEditor)
// document.querySelector("#lua-view").addEventListener("click", syncLuaEditor)
