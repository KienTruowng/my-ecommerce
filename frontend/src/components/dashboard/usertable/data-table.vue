<script setup lang="ts" generic="TData, TValue">
import type { ColumnDef, PaginationState } from '@tanstack/vue-table'
import { FlexRender, getCoreRowModel, useVueTable } from '@tanstack/vue-table'
import { Button } from '@/components/ui/button';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'

const props = defineProps<{
  columns: ColumnDef<TData, any>[]
  data: TData[]
  pageCount: number // Lấy từ api.meta.total_pages
  pagination: PaginationState // Sync state (pageIndex, pageSize) với cha
}>()

const emit = defineEmits<{
  'update:pagination': [updater: PaginationState]
}>()

const table = useVueTable({
  get data() { return props.data },
  get columns() { return props.columns },
  getCoreRowModel: getCoreRowModel(),

  // === SERVER-SIDE PAGINATION CONFIG ===
  manualPagination: true,
  get pageCount() { return props.pageCount },
  state: {
    get pagination() { return props.pagination }
  },
  onPaginationChange: (updaterOrValue) => {
    const newState =
      typeof updaterOrValue === 'function'
        ? updaterOrValue(props.pagination)
        : updaterOrValue
    emit('update:pagination', newState)
  }
})
</script>

<template>
  <div>
    <div class="border rounded-md">
      <Table>
        <TableHeader>
          <TableRow v-for="headerGroup in table.getHeaderGroups()" :key="headerGroup.id">
            <TableHead v-for="header in headerGroup.headers" :key="header.id">
              <FlexRender v-if="!header.isPlaceholder" :render="header.column.columnDef.header"
                :props="header.getContext()" />
            </TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <template v-if="table.getRowModel().rows?.length">
            <TableRow v-for="row in table.getRowModel().rows" :key="row.id"
              :data-state="row.getIsSelected() ? 'selected' : undefined">
              <TableCell v-for="cell in row.getVisibleCells()" :key="cell.id">
                <FlexRender :render="cell.column.columnDef.cell" :props="cell.getContext()" />
              </TableCell>
            </TableRow>
          </template>
          <template v-else>
            <TableRow>
              <TableCell :colspan="columns.length" class="h-24 text-center"> No results. </TableCell>
            </TableRow>
          </template>
        </TableBody>
      </Table>
    </div>

    <div class="flex items-center justify-center py-4 space-x-2">
      <Button variant="outline" size="lg" :disabled="!table.getCanPreviousPage()" @click="table.previousPage()">
        Previous
      </Button>
      <Button variant="outline" size="lg" :disabled="!table.getCanNextPage()" @click="table.nextPage()">
        Next
      </Button>
    </div>
  </div>
</template>