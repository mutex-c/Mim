# MIM 原神主题包设计规范文档 (Genshin Impact Theme Pack Design Spec)

> **创建日期**: 2026-10-25  
> **主题代号**: `genshin`（提瓦特星穹 · 原神）  
> **适用平台**: Desktop & Web (`clients/desktop-web`)  
> **设计系统准则**: 契合 MIM 现代轻质感设计系统，彻底杜绝粗糙低劣的 AI 拼凑感，打造精致、耐看、深度沉浸的原神专属视觉与交互体验。

---

## 1. 概述与设计目标

为 MIM 即时通讯客户端提供一套**全套沉浸式的原神专属主题包**，包含以下核心组成部分：
1. **提瓦特双模色盘 (`genshin`)**：
   - **深色模式（提瓦特夜空与星穹）**：祈愿夜空深蓝底色 + 5星愿望流金色 + 原石晶蓝高光 + 气泡微光金边；
   - **浅色模式（蒙德晨曦羊皮纸）**：象牙白与古籍暖沙底色 + 晨曦暖金 + 须弥青风点缀。
2. **星轨与星图氛围背景**：
   - 聊天主舞台在激活原神主题时呈现提瓦特几何星图与微尘星芒背景（纯矢量内嵌，零卡顿、无外链依赖）；
   - 当前会话项与激活状态带原石 `✦` 星芒徽记。
3. **原神角色预设身份库 (Genshin Character Personas)**：
   - 内置 8 位经典角色专属矢量头像、称号与预设资料（派蒙、钟离、芙宁娜、纳西妲、雷电将军、胡桃、温迪、万叶）；
   - 在个人资料修改中支持一键选择原神预设头像。
4. **原神名梗台词与表情贴纸系统 (Genshin Quotes & Stickers)**：
   - 在聊天输入框集成原神表情/贴纸抽屉面板；
   - 提供按元素（风、岩、雷、草、水、火、冰及通用名梗）分类的经典台词贴纸，一键发送。

---

## 2. 调色盘与 CSS 变量规范

### 2.1 CSS 变量映射表 (`clients/desktop-web/src/style.css`)

```css
/* ============================================================
   PALETTE 4: 提瓦特星穹 · 原神 (Teyvat Starlight & Genshin Gold)
   ============================================================ */

/* 浅色模式：蒙德晨曦 & 须弥日光 */
:root[data-palette='genshin'] {
  --mim-bg-dock: #f5f2e8;
  --mim-bg-list: #faf8f2;
  --mim-bg-stage: #eee9dc;
  --mim-bg-card: #ffffff;
  --mim-border: #e2d9c4;
  --mim-text: #26211a;
  --mim-muted: #786f62;
  --mim-hover: rgba(196, 139, 38, 0.06);

  --mim-accent: #c48b26;            /* 愿望流金 */
  --mim-accent-hover: #b27a1e;
  --mim-accent-subtle: rgba(196, 139, 38, 0.14);
  --mim-accent-text: #b27a1e;

  --mim-bubble-self: #c48b26;
  --mim-bubble-self-text: #ffffff;
  --mim-bubble-peer: #ffffff;
  --mim-bubble-peer-text: #26211a;

  --mim-group-badge: rgba(68, 168, 192, 0.16); /* 原石青 */
  --mim-group-badge-text: #1d7285;

  font-weight: 400;
  color-scheme: light;
}

/* 深色模式：提瓦特深邃星穹 & 祈愿金光 */
.dark[data-palette='genshin'],
:root[data-theme='dark'][data-palette='genshin'] {
  --mim-bg-dock: #0a0e19;
  --mim-bg-list: #0f1524;
  --mim-bg-stage: #0c111e;
  --mim-bg-card: #161e33;
  --mim-border: rgba(229, 176, 68, 0.20);
  --mim-text: #f0f4fc;
  --mim-muted: #828ea6;
  --mim-hover: rgba(229, 176, 68, 0.08);

  --mim-accent: #e5b044;            /* 5星金色传说 */
  --mim-accent-hover: #f5c563;
  --mim-accent-subtle: rgba(229, 176, 68, 0.22);
  --mim-accent-text: #f5c563;

  --mim-bubble-self: #19233c;       /* 带有金色微光边框 */
  --mim-bubble-self-text: #fef8eb;
  --mim-bubble-peer: #131b2e;
  --mim-bubble-peer-text: #f0f4fc;

  --mim-group-badge: rgba(88, 196, 220, 0.22); /* 原石青 */
  --mim-group-badge-text: #58c4dc;

  font-weight: 360;
  color-scheme: dark;
}
```

### 2.2 原神专属视觉修饰

在 `style.css` 中增加特定于 `[data-palette='genshin']` 的修饰符：
1. **星夜微尘背景**：为聊天舞台 `.chat-stage-bg` 叠加精致的 SVG 几何星穹渐变层；
2. **金色微光消息边框**：在深色模式下，发送方气泡增加 1px 的 `rgba(229, 176, 68, 0.35)` 微光金边与柔和辉光；
3. **原石四角星标志**：在当前激活会话标题旁呈现 `✦` 金色晶芒。

---

## 3. 角色预设与头像资产 (`src/assets/genshin/characters.ts`)

提供 8 位原神角色的预设数据结构与原生内嵌高品质 SVG 头像：
1. **派蒙 (Paimon)**：称号「最好的伙伴 / 应急食品」，主题色 `#e5b044`
2. **钟离 (Zhongli)**：称号「往生堂客卿 / 岩王帝君」，主题色 `#cfa858`
3. **芙宁娜 (Furina)**：称号「众水与戏剧的神明」，主题色 `#4ba6f5`
4. **纳西妲 (Nahida)**：称号「小吉祥草王」，主题色 `#76c867`
5. **雷电将军 (Raiden Shogun)**：称号「一心净土 / 鸣神大社」，主题色 `#9d6af5`
6. **胡桃 (Hu Tao)**：称号「往生堂第七十七代堂主」，主题色 `#e85642`
7. **温迪 (Venti)**：称号「风之诗人 / 巴巴托斯」，主题色 `#4fcda6`
8. **枫原万叶 (Kaedehara Kazuha)**：称号「浪迹天涯的浪人」，主题色 `#d96b52`

每个角色均携带高精度 SVG 图标，用户可在资料修改弹窗（`UserProfileModal.vue`）中一键切换头像与昵称。

---

## 4. 原神贴纸与名梗台词系统 (`src/assets/genshin/stickers.ts`)

### 4.1 数据结构
```ts
export interface GenshinSticker {
  id: string;
  character: string;
  element: 'Anemo' | 'Geo' | 'Electro' | 'Dendro' | 'Hydro' | 'Pyro' | 'Cryo' | 'Universal';
  quote: string;
  badge: string;
}
```

### 4.2 经典收录台词
- **通用/抽卡**：
  - “金色传说！⭐⭐⭐⭐⭐”
  - “原石 +1600！✦”
  - “应急食品，不是说好的最好的伙伴吗？”
  - “前面的区域，以后再来探索吧！”
- **岩系 (钟离)**：
  - “欲买桂花同载酒，只可惜故人，何日再见呢。”
  - “天动万象，此乃天理！”
- **水系 (芙宁娜/那维莱特)**：
  - “喝彩吧！沉醉吧！属于我的盛大剧目！”
  - “潮水啊，我已归来！”
- **雷系 (雷电将军)**：
  - “此身，当为稻妻之极致！”
  - “无念，断绝！”
- **草系 (纳西妲)**：
  - “此即，智慧之殿堂！”
- **火系 (胡桃)**：
  - “大丘丘病了二丘丘瞧，三丘丘采药四丘丘熬～”
- **风系 (温迪/万叶)**：
  - “起飞咯～”
  - “飘摇风雨中，带刀归来人。”

### 4.3 交互实现
在 `ChatStage.vue` 输入栏按钮组中新增 `GenshinStickerPicker` 弹窗。点击任一台词贴纸，直接填入输入框或作为格式化台词消息直接发送。

---

## 5. 组件集成与改造范围

1. **`clients/desktop-web/src/stores/theme.ts`**：
   - 拓展 `ColorPalette` 类型包含 `'genshin'`；
   - 在 `PALETTES` 数组中追加原神条目。
2. **`clients/desktop-web/src/style.css`**：
   - 追加原神主题 CSS 变量与专属星光视觉规则。
3. **`clients/desktop-web/src/components/SettingsModal.vue`**：
   - 色彩网格动态适配第 4 款主题，显示愿望金色彩与原神专属描述。
4. **`clients/desktop-web/src/components/ChatStage.vue`**：
   - 挂载原神贴纸抽屉面板；
   - 舞台背景与原神主题星轨样式适配。
5. **`clients/desktop-web/src/components/UserProfileModal.vue`**：
   - 增加“原神角色预设头像”选择区域，点击可快速应用。

---

## 6. 测试与验证计划

1. **色彩与主题切换测试**：
   - 在设置中切换到「提瓦特星穹 · 原神」；
   - 切换系统深色/浅色模式，验证两种模式下的对比度与可读性；
   - 验证刷新页面后持久化依然生效。
2. **角色头像应用测试**：
   - 打开用户资料，选取派蒙/钟离等角色预设，点击保存；
   - 验证左下角个人栏、聊天气泡头像均正确更新。
3. **名梗贴纸发送测试**：
   - 打开原神贴纸抽屉，点击“天动万象！”发送；
   - 验证对端实时收到并正确渲染。
4. **构建校验**：
   - 运行 `npm run build`，确保 TypeScript 类型与 Vite 打包 0 错误。
