import { autocompletion, closeBrackets, closeBracketsKeymap, completionKeymap } from "@codemirror/autocomplete"
import { defaultKeymap, history, historyKeymap, indentWithTab } from "@codemirror/commands"
import { bracketMatching, defaultHighlightStyle, foldGutter, foldKeymap, indentOnInput, syntaxHighlighting } from "@codemirror/language"
import { lintKeymap } from "@codemirror/lint"
import { highlightSelectionMatches, searchKeymap } from "@codemirror/search"
import { EditorState } from "@codemirror/state"
import { crosshairCursor, drawSelection, dropCursor, EditorView, highlightActiveLine, highlightActiveLineGutter, highlightSpecialChars, keymap, lineNumbers, rectangularSelection } from "@codemirror/view"
import { coolGlow } from "thememirror"
import { json } from "@codemirror/lang-json"

const jsonSource = document.querySelector("#json-editarea")
const jsonTarget = document.querySelector("#json-editor")

// Taken from: https://github.com/codemirror/basic-setup
// Mostly, at least. Also added small bits like the sibling tab-indent handling
const jsonSetup = [
  lineNumbers(),
  highlightActiveLineGutter(),
  highlightSpecialChars(),
  history(),
  foldGutter(),
  drawSelection(),
  dropCursor(),
  EditorState.allowMultipleSelections.of(true),
  indentOnInput(),
  syntaxHighlighting(defaultHighlightStyle, { fallback: true }),
  bracketMatching(),
  closeBrackets(),
  autocompletion(),
  rectangularSelection(),
  crosshairCursor(),
  highlightActiveLine(),
  highlightSelectionMatches(),
  keymap.of([
    ...closeBracketsKeymap,
    ...defaultKeymap,
    ...searchKeymap,
    ...historyKeymap,
    ...foldKeymap,
    ...completionKeymap,
    ...lintKeymap,
    indentWithTab
  ])
];

// Messing with state seemed to be trouble
// Mostly took this config from https://codemirror.net/examples/config/, near the end
let jsonView = new EditorView({
  doc: jsonSource.value,
  extensions: [
    jsonSetup,
    coolGlow,
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
