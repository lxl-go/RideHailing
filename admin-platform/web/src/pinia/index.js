import { createPinia } from 'pinia'
import { useAppStore } from '@/pinia/modules/app'
import { useThemeStore } from '@/pinia/modules/theme'
import { useUserStore } from '@/pinia/modules/user'
import { useDictionaryStore } from '@/pinia/modules/dictionary'
import { useOrderAdminStore } from '@/pinia/modules/order'
import { useLocationAdminStore } from '@/pinia/modules/location'

const store = createPinia()

export { store, useAppStore, useThemeStore, useUserStore, useDictionaryStore, useOrderAdminStore, useLocationAdminStore }
