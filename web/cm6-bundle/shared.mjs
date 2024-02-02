import { autocompletion, closeBrackets, closeBracketsKeymap, completionKeymap } from "@codemirror/autocomplete"
import { defaultKeymap, history, historyKeymap, indentWithTab } from "@codemirror/commands"
import { bracketMatching, defaultHighlightStyle, foldGutter, foldKeymap, indentOnInput, syntaxHighlighting } from "@codemirror/language"
import { lintKeymap } from "@codemirror/lint"
import { highlightSelectionMatches, searchKeymap } from "@codemirror/search"
import { EditorState } from "@codemirror/state"
import { crosshairCursor, drawSelection, dropCursor, EditorView, highlightActiveLine, highlightActiveLineGutter, highlightSpecialChars, keymap, lineNumbers, rectangularSelection } from "@codemirror/view"
import { coolGlow } from "thememirror"

// Taken from: https://github.com/codemirror/basic-setup
// Mostly, at least. Also added small bits like the sibling tab-indent handling
export const defaultSetup = [
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
  ]),
  coolGlow
]

// This could potentially turn into a performance problem for large scripts
// Does a good bit of string concat under the hood, revisit if needed
// Reference for this solution: https://discuss.codemirror.net/t/codemirror-6-proper-way-to-listen-for-changes/2395
// Docs to use changeset to more granularly update target value if needed: https://codemirror.net/docs/ref/#state.ChangeSet
export const syncEditorValue = (target) => {
  return EditorView.updateListener.of((v) => {
    if (v.docChanged) {
      target.value = v.state.doc.toString()
    }
  })
}
