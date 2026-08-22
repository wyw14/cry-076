import { beforeEach, describe, expect, it } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useWorkspaceStore } from './workspace'

function memoryStorage(): Storage {
  const values = new Map<string, string>()
  return {
    get length() { return values.size },
    clear: () => values.clear(),
    getItem: (key) => values.get(key) ?? null,
    key: (index) => [...values.keys()][index] ?? null,
    removeItem: (key) => { values.delete(key) },
    setItem: (key, value) => { values.set(key, value) },
  }
}

describe('workspace store', () => {
  beforeEach(() => {
	Object.defineProperty(window, 'localStorage', { configurable: true, value: memoryStorage() })
    window.localStorage.clear()
    setActivePinia(createPinia())
  })

  it('keeps draft fields when switching templates', () => {
    const store = useWorkspaceStore()
    const before = JSON.parse(JSON.stringify(store.draft))
    store.selectTemplate('focused')
    expect(store.selectedTemplate.id).toBe('focused')
    expect(store.draft).toEqual(before)
  })

  it('exposes only selected private fields', () => {
    const store = useWorkspaceStore()
    expect(store.selectedPrivateFields).toContain('email')
    expect(store.selectedPrivateFields).not.toContain('phone')
    store.togglePrivacy('phone')
    expect(store.selectedPrivateFields).toContain('phone')
  })
})
