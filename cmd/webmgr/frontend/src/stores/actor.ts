export const useActorStore = defineStore('actor', () => {
  const followedActors = ref<ActorDTO[]>([])
  const loading = ref(false)
  const followCount = ref(0)

  async function loadFollows() {
    loading.value = true
    try {
      const res = await api.getFollows()
      followedActors.value = res.data || []
    } finally {
      loading.value = false
    }
  }

  async function loadFollowCount() {
    const res = await api.getFollowCount()
    followCount.value = res.data
  }

  async function toggleFollow(actor: ActorDTO, isFollowed: boolean) {
    if (isFollowed) {
      await api.unfollowActor(actor.provider, actor.id)
    } else {
      await api.followActor(actor.provider, actor.id)
    }
    await loadFollows()
    await loadFollowCount()
  }

  return {
    followedActors,
    loading,
    followCount,
    loadFollows,
    loadFollowCount,
    toggleFollow,
  }
})
