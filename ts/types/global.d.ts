// Global type declarations

declare global {
  interface Window {
    _editorJumpToLine?: (lineNumber: number) => void
    dataLayer: unknown[]
    gtag: (...args: unknown[]) => void
  }
}

// This export is needed to make the file a module
export {}
