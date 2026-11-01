# 提瓦特星穹 · 原神全套主题包实施计划 (Genshin Impact Theme Pack Implementation Plan)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为 MIM 桌面与 Web 客户端打造全套高沉浸度原神主题包，涵盖深浅双模调色盘、提瓦特星空星轨视觉、8位经典角色预设头像库以及原神名梗与经典台词贴纸抽屉面板。

**Architecture:** 拓展 Pinia `useThemeStore` 色彩方案体系，通过 CSS 变量在 `style.css` 中注入提瓦特专属色彩与微光气泡层；抽离独立的原神角色库 (`characters.ts`) 与名梗贴纸库 (`stickers.ts`)；在 `ChatStage.vue` 中挂载原神贴纸抽屉与星轨背景，并在 `UserProfileModal.vue` 中集成一键原神换装能力。

**Tech Stack:** Vue 3 / TypeScript / Tailwind CSS / Pinia / Lucide Icons / SVG Vector Graphics

---

### File Structure Map

- **Modify**: [`clients/desktop-web/src/stores/theme.ts`](file:///e:/Work/Project/Mim/clients/desktop-web/src/stores/theme.ts) - 增加 `genshin` 调色盘类型与元数据
- **Modify**: [`clients/desktop-web/src/style.css`](file:///e:/Work/Project/Mim/clients/desktop-web/src/style.css) - 定义原神深浅双模变量、星夜背景与金色微光气泡
- **Create**: [`clients/desktop-web/src/assets/genshin/characters.ts`](file:///e:/Work/Project/Mim/clients/desktop-web/src/assets/genshin/characters.ts) - 8 位原神角色的预设数据与矢量 SVG 头像
- **Create**: [`clients/desktop-web/src/assets/genshin/stickers.ts`](file:///e:/Work/Project/Mim/clients/desktop-web/src/assets/genshin/stickers.ts) - 原神元素经典台词与名梗贴纸定义
- **Create**: [`clients/desktop-web/src/components/GenshinStickerPicker.vue`](file:///e:/Work/Project/Mim/clients/desktop-web/src/components/GenshinStickerPicker.vue) - 原神贴纸抽屉交互组件
- **Modify**: [`clients/desktop-web/src/components/ChatStage.vue`](file:///e:/Work/Project/Mim/clients/desktop-web/src/components/ChatStage.vue) - 挂载贴纸抽屉入口、原石星芒徽记与星轨背景
- **Modify**: [`clients/desktop-web/src/components/UserProfileModal.vue`](file:///e:/Work/Project/Mim/clients/desktop-web/src/components/UserProfileModal.vue) - 增加原神角色预设头像选择栏
- **Modify**: [`clients/desktop-web/src/components/SettingsModal.vue`](file:///e:/Work/Project/Mim/clients/desktop-web/src/components/SettingsModal.vue) - 调色盘数量动态化与原神卡片展示

---

### Task 1: 基础调色盘与全局 CSS 变量系统落地

**Files:**
- Modify: `clients/desktop-web/src/stores/theme.ts`
- Modify: `clients/desktop-web/src/style.css`

- [ ] **Step 1: 拓展 `theme.ts` 中的调色盘类型与列表**
  - 在 `ColorPalette` 类型中追加 `'genshin'`；
  - 在 `PALETTES` 数组中增加原神配置项：
    - `id: 'genshin'`
    - `nameZh: '提瓦特星穹 · 原神'`
    - `nameEn: 'Teyvat Starlight · Genshin'`
    - `previewColor: '#e5b044'`
    - `descZh: '提瓦特浩瀚星穹、原石晶蓝与愿望金辉，沉浸式幻想质感'`
    - `descEn: 'Teyvat deep celestial starlight with Primogem cyan and Wish gold accents'`

- [ ] **Step 2: 在 `style.css` 中注入原神深浅双模色彩与气泡微光样式**
  - 为 `:root[data-palette='genshin']` 编写浅色晨曦羊皮纸调色；
  - 为 `.dark[data-palette='genshin']` 与 `:root[data-theme='dark'][data-palette='genshin']` 编写深渊星穹深蓝、愿望流金色调；
  - 添加 `[data-palette='genshin'] .chat-stage-bg` 提瓦特几何星图背景规则；
  - 为深色模式下的 `.bubble-self` 添加微光金色边框与金色阴影；
  - 运行 `npx vue-tsc --noEmit` 验证类型正确。

---

### Task 2: 原神角色预设库与名梗贴纸资产开发

**Files:**
- Create: `clients/desktop-web/src/assets/genshin/characters.ts`
- Create: `clients/desktop-web/src/assets/genshin/stickers.ts`

- [ ] **Step 1: 编写 `characters.ts` 角色预设与 SVG 资产**
  - 定义 `GenshinCharacter` 接口（id、name、title、element、accentColor、avatarSvg）；
  - 编写 8 位经典角色专属的高品质矢量 SVG 头像：
    1. 派蒙 (Paimon) - 最好伙伴/应急食品
    2. 钟离 (Zhongli) - 往生堂客卿/岩王帝君
    3. 芙宁娜 (Furina) - 众水之神
    4. 纳西妲 (Nahida) - 小吉祥草王
    5. 雷电将军 (Raiden Shogun) - 稻妻神明
    6. 胡桃 (Hu Tao) - 往生堂堂主
    7. 温迪 (Venti) - 风之诗人
    8. 枫原万叶 (Kaedehara Kazuha) - 浪人武士
  - 导出 `GENSHIN_CHARACTERS` 数组。

- [ ] **Step 2: 编写 `stickers.ts` 元素名梗与台词贴纸库**
  - 定义 `GenshinSticker` 接口（id, character, element, quote, iconColor）；
  - 收录抽卡、大招语音、名台词等共计 16+ 条经典台词贴纸：
    - “金色传说！⭐⭐⭐⭐⭐”
    - “原石 +1600！✦”
    - “应急食品，不是说好的最好的伙伴吗？”
    - “天动万象，此乃天理！”
    - “欲买桂花同载酒，只可惜故人...”
    - “喝彩吧！沉醉吧！属于我的剧目！”
    - “此身，当为稻妻之极致！”
    - “此即，智慧之殿堂！”
    - “起飞咯～”
    - “大丘丘病了二丘丘瞧～”等。

---

### Task 3: 原神贴纸抽屉组件与聊天舞台集成

**Files:**
- Create: `clients/desktop-web/src/components/GenshinStickerPicker.vue`
- Modify: `clients/desktop-web/src/components/ChatStage.vue`

- [ ] **Step 1: 创建 `GenshinStickerPicker.vue`**
  - 悬浮毛玻璃卡片（Backdrop blur），支持按元素标签筛选（全部、抽卡、岩、水、雷、风、火、草）；
  - 呈现带有角色微型头像与名梗文字的质感卡片；
  - 点击贴纸触发 `select-sticker` 事件。

- [ ] **Step 2: 在 `ChatStage.vue` 中挂载原神贴纸抽屉**
  - 在输入栏工具区（图片、文件、麦克风旁边）增加原神原石星芒图标按钮（`Sparkles` 或专属贴纸图标）；
  - 激活时弹出 `GenshinStickerPicker`；
  - 选中贴纸时直接调用发送消息逻辑，生成高格式化台词消息；
  - 聊天背景与原神主题星轨样式适配。

---

### Task 4: 个人资料与设置面板的原神换装适配

**Files:**
- Modify: `clients/desktop-web/src/components/UserProfileModal.vue`
- Modify: `clients/desktop-web/src/components/SettingsModal.vue`

- [ ] **Step 1: 在 `UserProfileModal.vue` 中提供原神角色一键换装**
  - 在头像预览下方增加“原神角色预设”横向滚动手册；
  - 点击任意角色即可快速替换当前头像预览与昵称建议；
  - 点击保存即更新到用户 Profile。

- [ ] **Step 2: 在 `SettingsModal.vue` 中优化调色盘卡片**
  - 将原死循环的 `3 Sets Available` 改为动态 `{{ themeStore.palettes.length }} Sets Available`；
  - 确保原神金色主题卡片在栅格中呈现 5 星流金光泽。

---

### Task 5: 整体编译构建与全流程验证

**Files:**
- Verify: `clients/desktop-web`

- [ ] **Step 1: 前端代码静态检查与打包构建**
  - 执行 `npm run build`，验证 TypeScript 类型检查与 Vite 生产构建 0 错误。

- [ ] **Step 2: 完整用户链路验证**
  - 打开设置，切换至「提瓦特星穹 · 原神」；
  - 切换深色/浅色模式，验证配色准确度；
  - 打开个人资料，点击更换为“钟离”头像与昵称并保存；
  - 打开会话，点击原石贴纸按钮，发送“天动万象！”；
  - 验证气泡金边与文字渲染完美。
