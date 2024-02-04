import { StreamLanguage } from "@codemirror/language"
import { EditorView } from "@codemirror/view"
import { EditorState } from "@codemirror/state"
import { lua } from "@codemirror/legacy-modes/mode/lua"
import { defaultSetup, syncEditorValue } from "./shared.mjs"

const luaSource = document.querySelector("#lua-editarea")
const luaTarget = document.querySelector("#lua-editor")

new EditorView({
  state: EditorState.create({
    doc: luaSource.value,
    extensions: [
      defaultSetup,
      StreamLanguage.define(lua),
      syncEditorValue(luaSource)
    ]
  }),
  parent: luaTarget,
})
