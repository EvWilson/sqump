import { autocompletion, closeBrackets, closeBracketsKeymap, completionKeymap } from "@codemirror/autocomplete"
import { defaultKeymap, history, historyKeymap, indentWithTab } from "@codemirror/commands"
import { bracketMatching, defaultHighlightStyle, foldGutter, foldKeymap, indentOnInput, syntaxHighlighting } from "@codemirror/language"
import { lintKeymap } from "@codemirror/lint"
import { highlightSelectionMatches, searchKeymap } from "@codemirror/search"
import { EditorState } from "@codemirror/state"
import { StreamLanguage } from "@codemirror/language"
import { crosshairCursor, drawSelection, dropCursor, EditorView, highlightActiveLine, highlightActiveLineGutter, highlightSpecialChars, keymap, lineNumbers, rectangularSelection } from "@codemirror/view"
import { coolGlow } from "thememirror"
import { lua } from "@codemirror/legacy-modes/mode/lua"

const luaSource = document.querySelector("#lua-editarea")
const luaTarget = document.querySelector("#lua-editor")

// Taken from: https://github.com/codemirror/basic-setup
// Mostly, at least. Also added small bits like the sibling tab-indent handling
const luaSetup = [
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
let luaView = new EditorView({
  doc: luaSource.value,
  extensions: [
    luaSetup,
    coolGlow,
    StreamLanguage.define(lua)
  ],
  parent: luaTarget,
})

// Little trick from https://discuss.codemirror.net/t/codemirror-6-and-textareas/2731/3
const syncLuaEditor = () => {
  luaSource.value = luaView.state.sliceDoc()
  console.log(luaSource.value)
}

document.querySelector("#lua-submit").addEventListener("click", syncLuaEditor)
// Should also sync on other buttons
document.querySelector("#lua-exec").addEventListener("click", syncLuaEditor)
document.querySelector("#lua-view").addEventListener("click", syncLuaEditor)
