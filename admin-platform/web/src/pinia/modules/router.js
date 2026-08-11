import { asyncRouterHandle } from '@/utils/asyncRouter'
import { emitter } from '@/utils/bus.js'
import { defineStore } from 'pinia'
import { computed, ref, watchEffect } from 'vue'
import pathInfo from '@/pathInfo.json'
import {useRoute} from "vue-router";
import {config} from "@/core/config.js";

const notLayoutRouterArr = []
const keepAliveRoutersArr = []
const nameMap = {}

const formatRouter = (routes, routeMap, parent) => {
  routes &&
    routes.forEach((item) => {
      item.parent = parent
      item.meta.btns = item.btns
      item.meta.hidden = item.hidden
      if (item.meta.defaultMenu === true) {
        if (!parent) {
          item = { ...item, path: `/${item.path}` }
          notLayoutRouterArr.push(item)
        }
      }
      routeMap[item.name] = item
      if (item.children && item.children.length > 0) {
        formatRouter(item.children, routeMap, item)
      }
    })
}

const KeepAliveFilter = (routes) => {
  routes &&
    routes.forEach((item) => {
      if (
        (item.children && item.children.some((ch) => ch.meta.keepAlive)) ||
        item.meta.keepAlive
      ) {
        const path = item.meta.path
        keepAliveRoutersArr.push(pathInfo[path])
        nameMap[item.name] = pathInfo[path]
      }
      if (item.children && item.children.length > 0) {
        KeepAliveFilter(item.children)
      }
    })
}

function buildRideHailingMenus() {
  return [
    {
      path: 'dashboard',
      name: 'RideHailingDashboard',
      meta: { title: '运营概览', icon: 'el-icon-s-data', keepAlive: false },
      component: 'view/rideHailing/dashboard/index.vue'
    },
    {
      path: 'trips',
      name: 'RideHailingTrips',
      meta: { title: '行程审核', icon: 'el-icon-tickets', keepAlive: false },
      component: 'view/rideHailing/trips/index.vue'
    },
    {
      path: 'orders',
      name: 'RideHailingOrders',
      meta: { title: '订单管理', icon: 'el-icon-s-order', keepAlive: false },
      component: 'view/rideHailing/orders/index.vue'
    },
    {
      path: 'drivers',
      name: 'RideHailingDrivers',
      meta: { title: '司机管理', icon: 'el-icon-s-custom', keepAlive: false },
      component: 'view/rideHailing/drivers/index.vue'
    },
    {
      path: 'passengers',
      name: 'RideHailingPassengers',
      meta: { title: '乘客管理', icon: 'el-icon-s-check', keepAlive: false },
      component: 'view/rideHailing/passengers/index.vue'
    },
    {
      path: 'audit',
      name: 'RideHailingAudit',
      meta: { title: '认证审核', icon: 'el-icon-circle-check', keepAlive: false },
      component: 'view/admin/workorder01/review.vue'
    },
    {
      path: 'audit/vehicle',
      name: 'RideHailingVehicleAudit',
      meta: { title: '车辆审核', icon: 'el-icon-truck', keepAlive: false },
      component: 'view/admin/workorder01/vehicle-review.vue'
    },
    {
      path: 'shuttle',
      name: 'RideHailingShuttle',
      meta: { title: '班车管理', icon: 'el-icon-bus', keepAlive: false },
      component: 'view/admin/workorder02/shuttle.vue'
    },
    {
      path: 'finance',
      name: 'RideHailingFinance',
      meta: { title: '财务管理', icon: 'el-icon-money', keepAlive: false },
      component: 'view/admin/workorder03/finance.vue'
    },
    {
      path: 'members',
      name: 'RideHailingMembers',
      meta: { title: '人员管理', icon: 'el-icon-user', keepAlive: false },
      component: 'view/admin/workorder05/person.vue'
    },
    {
      path: 'analytics',
      name: 'RideHailingAnalytics',
      meta: { title: '数据分析', icon: 'el-icon-data-analysis', keepAlive: false },
      component: 'view/rideHailing/workorder06/analytics/index.vue'
    },
    {
      path: 'marketing',
      name: 'RideHailingMarketing',
      meta: { title: '营销管理', icon: 'el-icon-present', keepAlive: false },
      component: 'view/rideHailing/workorder07/marketing/index.vue'
    },
    {
      path: 'dispatch',
      name: 'RideHailingDispatch',
      meta: { title: '派单中心', icon: 'el-icon-guide', keepAlive: false },
      component: 'view/rideHailing/workorder11/dispatch/index.vue'
    },
    {
      path: 'ai',
      name: 'RideHailingAI',
      meta: { title: 'AI 助手', icon: 'el-icon-chat-dot-round', keepAlive: false },
      component: 'view/rideHailing/workorder10/ai/index.vue'
    },
    {
      path: 'performance',
      name: 'RideHailingPerformance',
      meta: { title: '性能监控', icon: 'el-icon-odometer', keepAlive: false },
      component: 'view/rideHailing/workorder08/performance/index.vue'
    },
    {
      path: 'governance',
      name: 'RideHailingGovernance',
      meta: { title: '平台治理', icon: 'el-icon-setting', keepAlive: false },
      component: 'view/rideHailing/workorder09/gva/index.vue'
    }
  ]
}

export const useRouterStore = defineStore('router', () => {
  const keepAliveRouters = ref([])
  const asyncRouterFlag = ref(0)
  const setKeepAliveRouters = (history) => {
    const keepArrTemp = []
    
    keepArrTemp.push(...keepAliveRoutersArr)
    if (config.keepAliveTabs) {
      history.forEach((item) => {
        const routeInfo = routeMap[item.name]
        if (routeInfo && routeInfo.meta && routeInfo.meta.path) {
          const componentName = pathInfo[routeInfo.meta.path]
          if (componentName) {
            keepArrTemp.push(componentName)
          }
        }
        
        if (nameMap[item.name]) {
          keepArrTemp.push(nameMap[item.name])
        }
      })
    }
    keepAliveRouters.value = Array.from(new Set(keepArrTemp))
  }

  const handleKeepAlive = async (to) => {
    if (!to.matched.some((item) => item.meta.keepAlive)) return

    if (to.matched?.length > 2) {
      for (let i = 1; i < to.matched.length; i++) {
        const element = to.matched[i - 1]

        if (element.name === 'layout') {
          to.matched.splice(i, 1)
          await handleKeepAlive(to)
          continue
        }

        if (typeof element.components.default === 'function') {
          await element.components.default()
          await handleKeepAlive(to)
        }
      }
    }
  }


  const route = useRoute()

  emitter.on('setKeepAlive', setKeepAliveRouters)

  const asyncRouters = ref([])

  const topMenu = ref([])

  const leftMenu = ref([])

  const menuMap = {}

  const topActive = ref('')

  const setLeftMenu = (name) => {
    sessionStorage.setItem('topActive', name)
    topActive.value = name
    leftMenu.value = []
    if (menuMap[name]?.children) {
      leftMenu.value = menuMap[name].children
    }
    return menuMap[name]?.children
  }

  const findTopActive = (menuMap, routeName) => {
    for (let topName in menuMap) {
      const topItem = menuMap[topName];
      if (topItem.children?.some(item => item.name === routeName)) {
        return topName;
      }
      const foundName = findTopActive(topItem.children || {}, routeName);
      if (foundName) {
        return topName;
      }
    }
    return null;
  };

  watchEffect(() => {
    let topActive = sessionStorage.getItem('topActive')
    topMenu.value = [];
    asyncRouters.value[0]?.children.forEach((item) => {
      if (item.hidden) return
      menuMap[item.name] = item
      topMenu.value.push({ ...item, children: [] })
    })
    if (!topActive || topActive === 'undefined' || topActive === 'null') {
      topActive = findTopActive(menuMap, route.name);
    }
    setLeftMenu(topActive)
  })

  const routeMap = {}
  const SetAsyncRouter = async () => {
    asyncRouterFlag.value++
    const baseRouter = [
      {
        path: '/layout',
        name: 'layout',
        component: 'view/layout/index.vue',
        meta: {
          title: '底层layout'
        },
        children: []
      }
    ]
    const menus = buildRideHailingMenus()
    menus.push({
      path: 'reload',
      name: 'Reload',
      hidden: true,
      meta: {
        title: '',
        closeTab: true
      },
      component: 'view/error/reload.vue'
    })
    formatRouter(menus, routeMap)
    baseRouter[0].children = menus
    if (notLayoutRouterArr.length !== 0) {
      baseRouter.push(...notLayoutRouterArr)
    }
    asyncRouterHandle(baseRouter)
    KeepAliveFilter(menus)
    asyncRouters.value = baseRouter
    return true
  }

  // 根级菜单（一级菜单树）：各布局取菜单数据的单一入口，避免 asyncRouters[0].children 散点解构。
  const rootMenus = computed(() => asyncRouters.value[0]?.children || [])

  return {
    topActive,
    setLeftMenu,
    topMenu,
    leftMenu,
    rootMenus,
    asyncRouters,
    keepAliveRouters,
    asyncRouterFlag,
    SetAsyncRouter,
    routeMap,
    handleKeepAlive
  }
})
