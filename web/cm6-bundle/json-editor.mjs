import { EditorView } from "@codemirror/view"
import { EditorState } from "@codemirror/state"
import { json } from "@codemirror/lang-json"
import { defaultSetup, syncEditorValue } from "./shared.mjs"

const jsonSource = document.getElementById("json-editarea")
const jsonTarget = document.getElementById("json-editor")

new EditorView({
  state: EditorState.create({
    doc: jsonSource.value,
    extensions: [
      defaultSetup,
      json(),
      syncEditorValue(jsonSource)
    ]
  }),
  parent: jsonTarget,
})
