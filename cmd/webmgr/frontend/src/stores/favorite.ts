export const useFavoriteStore = defineStore('favorite', () => {
  const folders = ref<FavoriteFolderDTO[]>([])
  const activeFolder = ref<number>(1)
  const loading = ref(false)

  async function loadFolders() {
    loading.value = true
    try {
      const res = await api.getFavoriteFolders()
      folders.value = res.data || []
      if (folders.value.length > 0 && !folders.value.find(f => f.id === activeFolder.value)) {
        activeFolder.value = folders.value[0].id
      }
    } finally {
      loading.value = false
    }
  }

  async function createFolder(name: string) {
    const res = await api.createFavoriteFolder(name)
    await loadFolders()
    return res.data
  }

  async function updateFolder(id: number, name: string) {
    await api.updateFavoriteFolder(id, name)
    await loadFolders()
  }

  async function deleteFolder(id: number) {
    await api.deleteFavoriteFolder(id)
    await loadFolders()
  }

  return {
    folders,
    activeFolder,
    loading,
    loadFolders,
    createFolder,
    updateFolder,
    deleteFolder,
  }
})
