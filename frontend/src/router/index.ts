import { useAuthStore } from '@/stores/auth'
import { useUserStore } from '@/stores/user'
import HomeView from '@/views/HomeView.vue'
import LoginView from '@/views/LoginView.vue'

import Profile from '@/views/settings/Profile.vue'
import SignUpView from '@/views/SignUpView.vue'
import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      component: LoginView,
      name: 'Login',
      meta: {
        layout: 'auth',
      },
    },
    {
      path: '/signup',
      component: SignUpView,
      meta: {
        layout: 'auth',
      },
    },
    {
      path: '/',
      component: HomeView,
      name: 'Home',
      meta: {
        layout: 'default',
        requiresAuth: true,
      },
    },
    {
      path: '/inventory',
      component: () => import('@/views/InventoryView.vue'),
      name: 'Inventory',
      meta: {
        layout: 'default',
        requiresAuth: true,
      },
    },
    {
      path: '/offers',
      redirect: '/orders',
    },
    {
      path: '/orders',
      component: () => import('@/views/OrdersView.vue'),
      name: 'Orders',
      meta: {
        layout: 'default',
        requiresAuth: true,
      },
    },
    {
      path: '/cart',
      component: () => import('@/views/CartView.vue'),
      name: 'Cart',
      meta: {
        layout: 'default',
        requiresAuth: true,
      },
    },
    {
      path: '/categories',
      component: () => import('@/views/CategoriesView.vue'),
      name: 'Categories',
      meta: {
        layout: 'default',
        requiresAuth: true,
      },
    },
    {
      path: '/wallet/topup/:code',
      component: () => import('@/views/TopupView.vue'),
      name: 'WalletTopup',
      meta: {
        layout: 'default',
        requiresAuth: true,
      },
    },
    {
      path: '/wallet',
      component: () => import('@/views/WalletView.vue'),
      name: 'Wallet',
      meta: {
        layout: 'default',
        requiresAuth: true,
      },
    },
    {
      path: '/appointments',
      component: () => import('@/views/AppointmentsView.vue'),
      name: 'Appointments',
      meta: {
        layout: 'default',
        requiresAuth: true,
      },
    },
    {
      path: '/ai-chat',
      component: () => import('@/views/AiChatView.vue'),
      name: 'AiChat',
      meta: {
        layout: 'default',
        requiresAuth: true,
      },
    },
    {
      path: '/recommendations',
      component: () => import('@/views/RecommendationsView.vue'),
      name: 'Recommendations',
      meta: {
        layout: 'default',
        requiresAuth: true,
      },
    },
    {
      path: '/items/:id',
      component: () => import('@/views/ItemDetailView.vue'),
      name: 'ItemDetail',
      meta: {
        layout: 'default',
        requiresAuth: true,
      },
    },
    {
      path: '/items/:id/auction',
      component: () => import('@/views/BiddingView.vue'),
      name: 'ItemAuction',
      meta: {
        layout: 'default',
        requiresAuth: true,
      },
    },
    {
      path: '/settings',
      component: () => import('@/views/settings/SettingsView.vue'),
      name: 'UserSettings',
      meta: {
        layout: 'default',
        requiresAuth: true,
      },
      children: [
        {
          path: 'profile',
          name: 'SettingsProfile',
          component: Profile,
        },
        {
          path: 'change-password',
          name: 'ChangePassword',
          component: () => import('@/views/settings/ChangePassword.vue'),
        },
      ],
    },
    {
      path: '/dashboard',
      component: () => import('@/views/dashboard/index.vue'),
      meta: {
        layout: 'empty',
        requiresAuth: true,
        requiresAdmin: true,
      },
      children: [
        {
          path: '',
          component: () => import('@/views/dashboard/DashboardHome.vue'),
        },
        {
          path: 'users',
          component: () => import('@/views/dashboard/ManagerUser.vue'),
        },
        {
          path: 'locations',
          component: () => import('@/views/dashboard/ManagerLocation.vue'),
        },
        {
          path: 'shipping',
          component: () => import('@/views/dashboard/ManagerShipping.vue'),
        },
      ],
    },
  ],
})

router.beforeEach(async (to, _from) => {
  const auth = useAuthStore()
  const userStore = useUserStore()

  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return { name: 'Login' }
  }

  if (to.meta.requiresAdmin) {
    if (!userStore.profile && auth.isAuthenticated) {
      await userStore.fetchProfile()
    }
    if (!userStore.isAdmin) {
      return { name: 'Home' }
    }
  }

  if (to.name === 'Login' && auth.isAuthenticated) {
    return { name: 'Home' }
  }

  return true
})

export default router
