import { test, expect } from '@playwright/test'

const BASE_URL = process.env.BASE_URL || 'http://localhost:5173'

test.describe('实例信息显示功能', () => {
  test.beforeEach(async ({ page }) => {
    // 登录
    await page.goto(`${BASE_URL}/login`)
    await page.fill('input[placeholder*="用户名"], input[placeholder*="username"]', 'admin')
    await page.fill('input[placeholder*="密码"], input[placeholder*="password"]', 'admin123')
    await page.click('button[type="submit"]')
    await page.waitForURL(`${BASE_URL}/services`)
  })

  test('服务列表页应显示实例摘要信息', async ({ page }) => {
    // 等待服务卡片加载
    await page.waitForSelector('.service-card')

    // 检查是否有服务
    const serviceCards = await page.locator('.service-card').count()
    expect(serviceCards).toBeGreaterThan(0)

    // 检查第一个服务卡片是否显示实例统计
    const firstCard = page.locator('.service-card').first()
    await expect(firstCard).toContainText(/实例:\s+\d+\/\d+/)
  })

  test('服务详情页应显示实例折叠面板', async ({ page }) => {
    // 点击第一个服务卡片
    await page.click('.service-card:first-child')

    // 等待详情页加载
    await page.waitForURL(/\/services\/.+/)

    // 检查实例卡片是否存在
    await expect(page.locator('.instances-card')).toBeVisible()

    // 检查实例列表标题
    await expect(page.locator('.instances-card h3')).toContainText('实例列表')
  })

  test('实例折叠面板应展开显示详细信息', async ({ page }) => {
    // 导航到服务详情页
    await page.goto(`${BASE_URL}/services/math-service`)

    // 等待实例列表加载
    await page.waitForSelector('.el-collapse-item', { timeout: 10000 })

    // 点击展开第一个实例
    await page.click('.el-collapse-item__header:first-child')

    // 验证详细信息显示
    await expect(page.locator('.instance-detail')).toBeVisible()
    await expect(page.locator('.el-descriptions')).toBeVisible()

    // 验证关键字段存在
    await expect(page.locator('text=/实例 Key/')).toBeVisible()
    await expect(page.locator('text=/语言/')).toBeVisible()
    await expect(page.locator('text=/主机 IP/')).toBeVisible()
  })

  test('管理员应能看到实例控制按钮', async ({ page }) => {
    // 导航到服务详情页
    await page.goto(`${BASE_URL}/services/math-service`)

    // 等待实例列表加载
    await page.waitForSelector('.el-collapse-item', { timeout: 10000 })

    // 展开第一个实例
    await page.click('.el-collapse-item__header:first-child')

    // 检查控制按钮是否存在
    await expect(page.locator('.instance-controls')).toBeVisible()
    await expect(page.locator('text=停止')).toBeVisible()
    await expect(page.locator('text=重启')).toBeVisible()
    await expect(page.locator('text=删除')).toBeVisible()
  })

  test('离线实例的停止和重启按钮应禁用', async ({ page }) => {
    await page.goto(`${BASE_URL}/services/math-service`)
    await page.waitForSelector('.el-collapse-item', { timeout: 10000 })

    // 找到离线实例并展开
    const offlineItem = page.locator('.el-collapse-item').filter({ hasText: '离线' }).first()
    const count = await offlineItem.count()

    if (count > 0) {
      await offlineItem.locator('.el-collapse-item__header').click()

      // 验证停止和重启按钮被禁用
      const stopBtn = offlineItem.locator('button:has-text("停止")')
      const restartBtn = offlineItem.locator('button:has-text("重启")')

      await expect(stopBtn).toBeDisabled()
      await expect(restartBtn).toBeDisabled()
    }
  })

  test('在线实例的删除按钮应禁用', async ({ page }) => {
    await page.goto(`${BASE_URL}/services/math-service`)
    await page.waitForSelector('.el-collapse-item', { timeout: 10000 })

    // 找到在线实例并展开
    const onlineItem = page.locator('.el-collapse-item').filter({ hasText: '在线' }).first()
    const count = await onlineItem.count()

    if (count > 0) {
      await onlineItem.locator('.el-collapse-item__header').click()

      // 验证删除按钮被禁用
      const deleteBtn = onlineItem.locator('button:has-text("删除")')
      await expect(deleteBtn).toBeDisabled()
    }
  })

  test('实例刷新功能', async ({ page }) => {
    await page.goto(`${BASE_URL}/services/math-service`)

    // 点击刷新按钮
    await page.click('.instances-card button:has-text("刷新")')

    // 验证 loading 状态
    await expect(page.locator('.instances-card .is-loading')).toBeVisible()
  })
})
