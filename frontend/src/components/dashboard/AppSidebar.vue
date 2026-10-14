<script setup lang="ts">
import { ArrowBigDown, MapPin, Truck, UsersRound } from 'lucide-vue-next'
import {
  Sidebar,
  SidebarContent,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from '../ui/sidebar'
import NavMain from './NavMain.vue'
import SearchInput from './SearchInput.vue'

const data = {
  navMain: [
    {
      title: 'Users',
      url: 'users',
      icon: UsersRound,
    },
    {
      title: 'Locations',
      url: 'locations',
      icon: MapPin,
    },
    {
      title: 'Shipping',
      url: 'shipping',
      icon: Truck,
    },
  ],
}
</script>
<template>
  <Sidebar variant="inset" collapsible="icon">
    <SidebarHeader>
      <SidebarMenu>
        <SidebarMenuItem>
          <SidebarMenuButton
            as-child
            class="data-[slot=sidebar-menu-button]:p-1.5!"
          >
            <a href="#">
              <ArrowBigDown class="size-5!" />
              <span class="text-base font-semibold">CampusTrade</span>
            </a>
          </SidebarMenuButton>
        </SidebarMenuItem>
      </SidebarMenu>
    </SidebarHeader>
    <SidebarContent>
      <SearchInput />
      <NavMain :items="data.navMain" base-path="/dashboard" />
    </SidebarContent>
  </Sidebar>
</template>
