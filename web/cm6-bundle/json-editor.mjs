import { EditorView } from "@codemirror/view"
import { EditorState } from "@codemirror/state"
import { json } from "@codemirror/lang-json"
import { defaultSetup, syncEditorValue } from "./shared.mjs"

const jsonSource = document.querySelector("#json-editarea")
const jsonTarget = document.querySelector("#json-editor")

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
