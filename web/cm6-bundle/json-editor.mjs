import { EditorView } from "@codemirror/view"
import { json } from "@codemirror/lang-json"
import { defaultSetup } from "./shared.mjs"

const jsonSource = document.querySelector("#json-editarea")
const jsonTarget = document.querySelector("#json-editor")

// Messing with state seemed to be trouble
// Mostly took this config from https://codemirror.net/examples/config/, near the end
let jsonView = new EditorView({
  doc: jsonSource.value,
  extensions: [
    defaultSetup,
    json()
  ],
  parent: jsonTarget,
})

// Little trick from https://discuss.codemirror.net/t/codemirror-6-and-textareas/2731/3
const syncJsonEditor = () => {
  jsonSource.value = jsonView.state.sliceDoc()
  console.log(jsonSource.value)
}

const jsonSubmit = document.querySelector("#json-submit")
jsonSubmit.addEventListener("click", syncJsonEditor)
