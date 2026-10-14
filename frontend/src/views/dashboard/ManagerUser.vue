<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import DataTable from '@/components/dashboard/usertable/data-table.vue'
import type { paths } from '@/schemas/types'
import type { PaginationState } from '@tanstack/vue-table'

// 1. Thay đổi import: Import getColumns thay vì columns
import { getColumns } from '@/components/dashboard/usertable/columns'

// 2. Import các component của Shadcn UI
import {
    Sheet,
    SheetContent,
    SheetHeader,
    SheetTitle,
    SheetDescription,
    SheetFooter,
    SheetClose,
} from '@/components/ui/sheet'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { api } from '@/api/api'
import {
    Dialog,
    DialogClose,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger, // Dù không dùng trực tiếp ở đây nhưng vẫn giữ lại theo code gốc của bạn
} from '@/components/ui/dialog'
import { toast } from 'vue-sonner'

// --- TYPE DEFINITIONS ---
type UsersResponse =
    paths['/api/admin/users']['get']['responses']['200']['content']['application/json']
type UserList = NonNullable<UsersResponse['data']>
type User = UserList[number] // Lấy type của 1 user cụ thể

// --- STATES ---
const data = ref<UserList>([])

const isSheetOpen = ref(false)
const editingUser = ref<User | null>(null)
const deletingUserId = ref<number | null>(null)
const isDialogOpen = ref(false)

// --- PAGINATION STATES ---
const totalPages = ref(0)
const pagination = ref<PaginationState>({
    pageIndex: 0, // 0 tương đương với trang 1 (Tanstack table bắt đầu từ 0)
    pageSize: 10,
})

// --- HANDLERS ---
const handleEditUser = (user: User) => {
    editingUser.value = JSON.parse(JSON.stringify(user))
    isSheetOpen.value = true
}

const handleDeleteUser = (user: User) => {
    deletingUserId.value = user.id ?? null
    isDialogOpen.value = true
}

const onSubmit = async () => {
    if (!editingUser.value) return

    try {
        const { data: responseData, error } = await api.PATCH('/api/admin/users/{id}', {
            params: {
                path: {
                    id: editingUser.value.id,
                },
            },
            body: {
                name: editingUser.value.name,
                email: editingUser.value.email,
                role: editingUser.value.role,
                is_verified: editingUser.value.is_verified,
                university_id: editingUser.value.university?.id,
            },
        })

        if (error) {
            console.error('Lỗi khi lưu user:', error)
            toast.error('Không thể cập nhật user')
            return
        }

        const updatedUser = responseData?.data
        const index = data.value.findIndex((u) => u.id === editingUser.value?.id)
        if (index !== -1 && updatedUser) {
            data.value[index] = { ...data.value[index], ...updatedUser }
        }

        toast.success('Cập nhật user thành công')
        isSheetOpen.value = false
    } catch (error) {
        console.error('Lỗi khi lưu user:', error)
    }
}

// Callback
const tableColumns = getColumns(handleEditUser, handleDeleteUser)

// --- API FETCHING ---
async function getData(): Promise<UserList> {
    try {
        const { data: responseData, error } = await api.GET('/api/admin/users', {
            params: {
                query: {
                    page: pagination.value.pageIndex + 1, //  0-based sang 1-based
                    limit: pagination.value.pageSize,
                },
            },
        })

        if (error) {
            console.error('Lỗi từ API:', error)
            return []
        }


        if (responseData?.meta?.total_pages) {
            totalPages.value = responseData.meta.total_pages
        }

        return responseData?.data || []
    } catch (error) {
        console.error('Lỗi khi lấy danh sách user:', error)
        return []
    }
}

const deleteUser = async () => {
    if (!deletingUserId.value) return
    try {
        const { error } = await api.DELETE('/api/admin/users/{id}', {
            params: {
                path: {
                    id: deletingUserId.value,
                },
            },
        })

        if (error) {
            console.error('Lỗi khi xóa user:', error)
            toast.error('Xuất hiện lỗi khi xóa user')
            return
        }

        toast.success('Xóa thành công user')
        data.value = data.value.filter((u) => u.id !== deletingUserId.value)

        isDialogOpen.value = false
        deletingUserId.value = null

    } catch (err) {
        console.error('Lỗi không mong muốn khi xóa:', err)
    }
}

watch(
    pagination,
    async () => {
        data.value = await getData()
    },
    { deep: true }
)

onMounted(async () => {
    data.value = await getData()
})
</script>

<template>
    <div class="flex flex-1 flex-col">
        <div class="@container/main flex flex-1 flex-col gap-2">
            <div class="flex flex-col gap-4 py-4 md:gap-6 md:py-6">
                <div>
                    <DataTable :columns="tableColumns" :data="data" :pageCount="totalPages"
                        v-model:pagination="pagination" />
                </div>
            </div>
        </div>

        <Sheet v-model:open="isSheetOpen">
            <SheetContent>
                <SheetHeader>
                    <SheetTitle>Chỉnh sửa người dùng</SheetTitle>
                    <SheetDescription> Cập nhật thông tin cho tài khoản </SheetDescription>
                </SheetHeader>
                <form @submit.prevent="onSubmit" class="flex flex-1 flex-col">
                    <div class="grid flex-1 auto-rows-min gap-6 px-4">
                        <div v-if="editingUser" class="grid gap-4 py-6">
                            <div class="grid gap-4">
                                <Label for="name" class="text-right">Họ và tên</Label>
                                <Input id="name" v-model="editingUser.name" class="col-span-3" />
                            </div>

                            <div class="grid gap-4">
                                <Label for="email" class="text-right">Email</Label>
                                <Input id="email" v-model="editingUser.email" disabled
                                    class="col-span-3 bg-muted text-muted-foreground" />
                            </div>
                        </div>
                    </div>

                    <SheetFooter>
                        <Button type="submit"> Save changes </Button>
                        <SheetClose as-child>
                            <Button variant="outline"> Close </Button>
                        </SheetClose>
                    </SheetFooter>
                </form>
            </SheetContent>
        </Sheet>

        <Dialog v-model:open="isDialogOpen">
            <DialogContent>
                <form @submit.prevent="deleteUser">
                    <DialogHeader>
                        <DialogTitle>Are you absolutely sure?</DialogTitle>
                        <DialogDescription>
                            This action cannot be undone. This will permanently delete your account and remove
                            your data from our servers.
                        </DialogDescription>
                    </DialogHeader>

                    <DialogFooter>
                        <DialogClose as-child>
                            <Button variant="outline"> Cancel </Button>
                        </DialogClose>
                        <Button type="submit" variant="destructive"> Xác nhận xóa </Button>
                    </DialogFooter>
                </form>
            </DialogContent>
        </Dialog>
    </div>
</template>
