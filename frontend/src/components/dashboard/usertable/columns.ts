import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import type { paths } from '@/schemas/types'
import { createColumnHelper } from '@tanstack/vue-table'
import { h } from 'vue'

type UsersResponseData =
  paths['/api/admin/users']['get']['responses']['200']['content']['application/json']['data']

type User = NonNullable<UsersResponseData>[number]

const columnHelper = createColumnHelper<User>()

/**
 * Helper function for render header
 */
const renderHeader = (title: string, alignment: 'left' | 'center' | 'right' = 'left') => {
  const alignClass =
    alignment === 'right' ? 'text-right' : alignment === 'center' ? 'text-center' : 'text-left'

  return h(
    'div',
    {
      class: `text-[11px] font-bold uppercase tracking-widest text-muted-foreground/80 ${alignClass}`,
    },
    title,
  )
}

export const getColumns = (onEdit: (user: User) => void, onDelete: (user: User) => void) => [
  columnHelper.accessor('id', {
    header: () => renderHeader('ID'),
    cell: (info) => h('div', { class: 'font-mono text-xs text-gray-500' }, info.getValue()),
  }),

  columnHelper.accessor('avatar_url', {
    header: () => renderHeader('Ảnh'),
    cell: (info) => {
      const url = info.getValue()
      const userName = info.row.original.name || 'US'
      const fallbackText = userName.substring(0, 2).toUpperCase()

      return h(Avatar, { class: 'w-9 h-9 border border-gray-100' }, () => [
        h(AvatarImage, { src: url || '', alt: userName }),
        h(AvatarFallback, { class: 'bg-slate-100 text-xs' }, () => fallbackText),
      ])
    },
  }),

  columnHelper.accessor('name', {
    header: () => renderHeader('Họ và tên'),
    cell: (info) => h('div', { class: 'font-semibold text-gray-900' }, info.getValue() || 'N/A'),
  }),

  columnHelper.accessor('email', {
    header: () => renderHeader('Email'),
    cell: (info) => h('div', { class: 'text-gray-600' }, info.getValue() || 'N/A'),
  }),

  columnHelper.accessor('role', {
    header: () => renderHeader('Vai trò', 'center'),
    cell: (info) => {
      const role = info.getValue()
      const badgeClass =
        role === 'admin'
          ? 'bg-purple-100 text-purple-700 px-2 py-1 rounded-md text-[10px] font-bold border border-purple-200'
          : 'bg-blue-100 text-blue-700 px-2 py-1 rounded-md text-[10px] font-bold border border-blue-200'

      return h('div', { class: 'flex justify-center' }, [
        h('span', { class: badgeClass }, role ? role.toUpperCase() : 'USER'),
      ])
    },
  }),

  columnHelper.accessor('is_verified', {
    header: () => renderHeader('Xác minh', 'center'),
    cell: (info) => {
      const isVerified = info.getValue()
      return h('div', { class: 'flex justify-center' }, [
        h(
          Badge,
          {
            variant: isVerified ? 'default' : 'destructive',
            class: 'text-[10px] uppercase font-bold',
          },
          () => [isVerified ? '✓ Đã xác minh' : '✕ Chưa xác minh'],
        ),
      ])
    },
  }),

  columnHelper.accessor((row) => row.university?.name, {
    id: 'university',
    header: () => renderHeader('Trường Đại học'),
    cell: (info) =>
      h('div', { class: 'text-sm text-gray-600' }, info.getValue() || 'Chưa cập nhật'),
  }),

  columnHelper.accessor('created_at', {
    header: () => renderHeader('Ngày đăng ký'),
    cell: (info) => {
      const dateStr = info.getValue()
      if (!dateStr) return 'N/A'

      return h(
        'div',
        { class: 'text-gray-500 text-sm' },
        new Date(dateStr).toLocaleDateString('vi-VN', {
          day: '2-digit',
          month: '2-digit',
          year: 'numeric',
        }),
      )
    },
  }),

  columnHelper.display({
    id: 'actions',
    header: () => renderHeader('Hành động', 'right'),
    cell: ({ row }) => {
      const user = row.original

      return h('div', { class: 'flex justify-end items-center gap-2' }, [
        h(
          Button,
          {
            size: 'sm',
            class: 'h-8 px-3 text-xs',
            onClick: () => {
              onEdit(user) // onEdit bây giờ đã được nhận từ tham số của getColumns
            },
          },
          () => 'Sửa',
        ),
        h(
          Button,
          {
            variant: 'destructive',
            size: 'sm',
            class: 'h-8 px-3 text-xs',
            onClick: () => {
              onDelete(user)
            },
          },
          () => 'Xóa',
        ),
      ])
    },
  }),
]
