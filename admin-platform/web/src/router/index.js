import { createRouter, createWebHashHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    redirect: '/login'
  },
  {
    path: '/init',
    name: 'Init',
    component: () => import('@/view/init/index.vue')
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/view/login/index.vue')
  },
  {
    path: '/scanUpload',
    name: 'ScanUpload',
    meta: {
      title: 'Scan upload',
      client: true
    },
    component: () => import('@/view/media/scanUpload.vue')
  },
  {
    path: '/forceChangePassword',
    name: 'ForceChangePassword',
    component: () => import('@/view/system/security/forceChangePassword.vue'),
    meta: { title: 'Change password' }
  },
  {
    path: '/ride-hailing',
    name: 'RideHailing',
    component: () => import('@/view/layout/index.vue'),
    redirect: '/ride-hailing/dashboard',
    meta: { title: 'Ride hailing' },
    children: [
      {
        path: 'dashboard',
        name: 'RideHailingDashboard',
        component: () => import('@/view/rideHailing/dashboard/index.vue'),
        meta: { title: '运营概览' }
      },
      {
        path: 'trips',
        name: 'RideHailingTrips',
        component: () => import('@/view/rideHailing/trips/index.vue'),
        meta: { title: '行程审核' }
      },
      {
        path: 'orders',
        name: 'RideHailingOrders',
        component: () => import('@/view/rideHailing/orders/index.vue'),
        meta: { title: '订单管理' }
      },
      {
        path: 'drivers',
        name: 'RideHailingDrivers',
        component: () => import('@/view/rideHailing/drivers/index.vue'),
        meta: { title: '司机管理' }
      },
      {
        path: 'passengers',
        name: 'RideHailingPassengers',
        component: () => import('@/view/rideHailing/passengers/index.vue'),
        meta: { title: '乘客管理' }
      },
      {
        path: 'audit',
        name: 'RideHailingAudit',
        component: () => import('@/view/admin/workorder01/review.vue'),
        meta: { title: '认证审核' }
      },
      {
        path: 'audit/vehicle',
        name: 'RideHailingVehicleAudit',
        component: () => import('@/view/admin/workorder01/vehicle-review.vue'),
        meta: { title: '车辆审核' }
      },
      {
        path: 'shuttle',
        name: 'RideHailingShuttle',
        component: () => import('@/view/admin/workorder02/shuttle.vue'),
        meta: { title: '班车管理' }
      },
      {
        path: 'finance',
        name: 'RideHailingFinance',
        component: () => import('@/view/admin/workorder03/finance.vue'),
        meta: { title: '财务管理' }
      },
      {
        path: 'members',
        name: 'RideHailingMembers',
        component: () => import('@/view/admin/workorder05/person.vue'),
        meta: { title: '人员管理' }
      },
      {
        path: 'analytics',
        name: 'RideHailingAnalytics',
        component: () => import('@/view/rideHailing/workorder06/analytics/index.vue'),
        meta: { title: '数据分析' }
      },
      {
        path: 'marketing',
        name: 'RideHailingMarketing',
        component: () => import('@/view/rideHailing/workorder07/marketing/index.vue'),
        meta: { title: '营销管理' }
      },
      {
        path: 'dispatch',
        name: 'RideHailingDispatch',
        component: () => import('@/view/rideHailing/workorder11/dispatch/index.vue'),
        meta: { title: '派单中心' }
      },
      {
        path: 'ai',
        name: 'RideHailingAI',
        component: () => import('@/view/rideHailing/workorder10/ai/index.vue'),
        meta: { title: 'AI 助手' }
      },
      {
        path: 'performance',
        name: 'RideHailingPerformance',
        component: () => import('@/view/rideHailing/workorder08/performance/index.vue'),
        meta: { title: '性能监控' }
      },
      {
        path: 'governance',
        name: 'RideHailingGovernance',
        component: () => import('@/view/rideHailing/workorder09/gva/index.vue'),
        meta: { title: '平台治理' }
      }
    ]
  },
  {
    path: '/:catchAll(.*)',
    meta: {
      closeTab: true
    },
    component: () => import('@/view/error/index.vue')
  }
]

const router = createRouter({
  history: createWebHashHistory(),
  routes
})

export default router
