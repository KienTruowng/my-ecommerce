<script setup lang="ts">
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'

import {
  InputGroup,
  InputGroupAddon,
  InputGroupButton,
  InputGroupInput,
  InputGroupText,
  InputGroupTextarea,
} from '@/components/ui/input-group'
import { SearchIcon } from 'lucide-vue-next'
import { Kbd, KbdGroup } from '../ui/kbd'
import { useMagicKeys, whenever } from '@vueuse/core'
import { onMounted, ref } from 'vue'
const keys = useMagicKeys()

const ctrlSlash = keys['ctrl+/']
const metaSlash = keys['meta+/']

const isOpen = ref(false)
const isMac = ref(false)

onMounted(() => {
  if (typeof window !== 'undefined') {
    isMac.value = navigator.userAgent.toUpperCase().includes('MAC')
  }
})

whenever(
  () => ctrlSlash?.value || metaSlash?.value,
  () => {
    isOpen.value = !isOpen.value
  },
)
</script>

<template>
  <Dialog v-model:open="isOpen">
    <DialogTrigger as-child>
      <InputGroup>
        <InputGroupAddon align="inline-start">
          <SearchIcon />
        </InputGroupAddon>
        <InputGroupInput placeholder="Search..." readonly />
        <InputGroupAddon align="inline-end">
          <KbdGroup>
            <Kbd>{{ isMac ? '⌘' : 'Ctrl' }}</Kbd>
            <span>+</span>
            <Kbd>/</Kbd>
          </KbdGroup>
        </InputGroupAddon>
      </InputGroup>
    </DialogTrigger>

    <DialogContent :show-close-button="false">
      <DialogHeader>
        <InputGroup>
          <InputGroupAddon align="inline-start">
            <SearchIcon />
          </InputGroupAddon>
          <InputGroupInput placeholder="Search..." />
        </InputGroup>
      </DialogHeader>
    </DialogContent>
  </Dialog>
</template>
