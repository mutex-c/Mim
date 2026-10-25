# MIM 沃雅妮莎深度定制主题包设计规范 (Vodyanitsa Bespoke Theme Pack Design Spec)

> **创建日期**: 2026-10-25  
> **主题代号**: `vodyanitsa`（幽歌萦渊 · 沃雅妮莎）  
> **适用平台**: Desktop & Web (`clients/desktop-web`)  
> **设计准则**: 围绕《原神》至冬国科洛列夫茨基剧团首席女高音、离群水妖「沃雅妮莎」进行全方位深度量身定制。以「深渊剧院曜蓝」、「水妖空灵水晶青」与「科洛列夫茨基殿堂金」为主色系，打造媲美官方高定画作与剧情对话水准的沉浸式体验。

---

## 1. 概述与核心目标

本主题包为 MIM 单一角色深度定制的标杆之作，通过全方位的视觉、动效、资产与音效构建角色沉浸感：
1. **专属双模调色盘 (`vodyanitsa`)**：
   - **深色模式（深渊歌剧院曜夜）**：深邃曜蓝夜空（`#080f24` / `#040714`）+ 水妖水晶冰蓝（`#38bdf8` / `#0284c7`）+ 剧院殿堂流金（`#eab308` / `#fef08a`）；
   - **浅色模式（至冬冰泉晨曦）**：象牙银白（`#f8fafc`）+ 冰湖天青（`#0284c7`）+ 浮雕冷金点缀。
2. **剧情级歌剧院水妖对话底板 (Opera Siren Dialogue Plates)**：
   - **对方（沃雅妮莎 / Peer 消息）**：深海曜蓝石板材质、微光水波漫射、水之眼宝石切割徽章、`★★★★★` 金色流光星级、`✦ 幽歌萦渊 · 首席女高音` 专属称号、晶莹水滴微标分割线（`—— 💧 ——`）、四角水晶凯尔特包角；
   - **自己（旅行者 / Self 消息）**：曜石暖金星辉材质、璨金流光边框、`★★★★★` 星级、`✦ 异乡的旅人 · 首席乐迷` 专属头衔、金菱暗纹分割线、右侧专属「**旅**」金框头像。
3. **AI 原生专属艺术资产库 (Artworks & Chibi Stickers)**：
   - 1 张官方级五星祈愿大立绘（`splash_vodyanitsa.png`）；
   - 6 张专属微表情贴纸（`vodyanitsa_sing.png` 高歌咏叹、`vodyanitsa_pout.png` 委屈鼓脸、`vodyanitsa_horn.png` 唤春角笛、`vodyanitsa_dive.png` 探头冒泡、`vodyanitsa_tea.png` 剧院红茶、`vodyanitsa_bow.png` 傲娇谢幕）。
4. **「幽歌萦渊」专属限定祈愿抽卡系统**：
   - 专属活动卡池，十连抽金色流星与水波飞舞动效，出金全屏展示立绘与台词，一键分享专属水纹战报卡片至聊天。
5. **金鳍座 · 科洛列夫茨基剧团首席档案展柜**：
   - 个人资料弹窗深度定制：金鳍座（Aura Fin）命之座纹章、生命值 65,280、水冰抗性 -35%、治疗加成 +75.4% 战斗面板还原。

---

## 2. 调色盘与 CSS 变量规范

### 2.1 CSS 变量映射表 (`clients/desktop-web/src/style.css`)

```css
/* ============================================================
   PALETTE 5: 幽歌萦渊 · 沃雅妮莎 (Vodyanitsa Opera & Crystal Ocean)
   ============================================================ */

/* 浅色模式：至冬冰泉晨曦 */
:root[data-palette='vodyanitsa'] {
  --mim-bg-dock: #f1f5f9;
  --mim-bg-list: #f8fafc;
  --mim-bg-stage: #e2e8f0;
  --mim-bg-card: #ffffff;
  --mim-border: #cbd5e1;
  --mim-text: #0f172a;
  --mim-muted: #64748b;
  --mim-hover: rgba(2, 132, 199, 0.08);

  --mim-accent: #0284c7;            /* 水妖湖蓝 */
  --mim-accent-hover: #0369a1;
  --mim-accent-subtle: rgba(2, 132, 199, 0.12);
  --mim-accent-text: #0369a1;

  --mim-bubble-self: #0284c7;
  --mim-bubble-self-text: #ffffff;
  --mim-bubble-peer: #ffffff;
  --mim-bubble-peer-text: #0f172a;

  --mim-group-badge: rgba(56, 189, 248, 0.16);
  --mim-group-badge-text: #0284c7;
}

/* 深色模式：深渊歌剧院曜夜 */
.dark[data-palette='vodyanitsa'],
:root[data-theme='dark'][data-palette='vodyanitsa'] {
  --mim-bg-dock: #040817;
  --mim-bg-list: #080f24;
  --mim-bg-stage: #050b1a;
  --mim-bg-card: #0d1733;
  --mim-border: rgba(56, 189, 248, 0.18);
  --mim-text: #f0f9ff;
  --mim-muted: #7dd3fc;
  --mim-hover: rgba(56, 189, 248, 0.12);

  --mim-accent: #38bdf8;            /* 水晶冰蓝 */
  --mim-accent-hover: #7dd3fc;
  --mim-accent-subtle: rgba(56, 189, 248, 0.2);
  --mim-accent-text: #e0f2fe;

  --mim-bubble-self: #0284c7;
  --mim-bubble-self-text: #ffffff;
  --mim-bubble-peer: #0e172e;
  --mim-bubble-peer-text: #f0f9ff;

  --mim-group-badge: rgba(56, 189, 248, 0.22);
  --mim-group-badge-text: #7dd3fc;
}
```

---

## 3. 剧情对话底板与微表情贴纸组件设计

### 3.1 对话底板结构与样式规范

1. **对方底板（`.vodyanitsa-plate-peer`）**：
   - 材质背景：
     ```css
     background-color: #0b1329 !important;
     background: radial-gradient(circle at 10% 0%, rgba(56, 189, 248, 0.18) 0%, transparent 70%),
                 linear-gradient(175deg, #101935 0%, #060a16 100%) !important;
     border: 1.5px solid rgba(56, 189, 248, 0.7) !important;
     box-shadow: inset 0 1px 1px rgba(224, 242, 254, 0.45), 0 8px 24px rgba(0, 0, 0, 0.5) !important;
     ```
   - 顶栏：水元素切割神之眼（`水` 字体，泛浅蓝光芒）+ 角色名 `沃雅妮莎` + 流金 `★★★★★` + 称号 `✦ 幽歌萦渊 · 首席女高音`。
   - 分割线：细金冰蓝渐变线，中央缀晶莹水滴 `💧`。
   - 光标：右下角跳动的水晶 `✦` 光标。

2. **自己发出的底板（`.vodyanitsa-plate-self`）**：
   - 材质背景：深曜岩暖金石板材质（`#221a14` ~ `#0d0c14`），包角与边框为璨金流光（`rgba(234, 179, 8, 0.8)`）。
   - 顶栏：岩元素切割宝石徽章 + `旅行者` + `★★★★★` + `✦ 异乡的旅人 · 首席乐迷`。
   - 右侧头像：专属金框 `旅` 头像。

3. **微表情贴纸面板（`VodyanitsaStickerPicker.vue`）**：
   - 点击聊天工具栏「💧 沃雅妮莎」贴纸按钮展开贴纸抽屉；
   - 渲染 6 张由 AI 生成的高清水妖微表情图，支持悬浮预览和点击发送。
   - 聊天主舞台针对 `[VODYANITSA_STICKER]:<id>` 解析为精美水妖表情大卡片，不显示冗余代码块。

---

## 4. 专属祈愿卡池与角色展柜设计

### 4.1 专属限定祈愿「幽歌萦渊」(`VodyanitsaWishModal.vue`)
- **卡池海报**：沃雅妮莎官方级大立绘，背景流淌深渊夜空与水流乐符。
- **十连抽机制**：
  - 模拟真实原神抽卡概率与十连出金保底；
  - 抽出 5 星沃雅妮莎时触发金色大流星与华丽立绘定格，音效联动；
  - 底部提供「✦ 一键分享战报」按钮，将带有水妖金蓝边框的祈愿成果卡片直接投递至聊天舞台。

### 4.2 个人资料「金鳍座」档案展柜 (`UserProfileModal.vue`)
- 切换至 `vodyanitsa` 主题时，个人资料展示沃雅妮莎专属名片与立绘展柜；
- 显示角色专属称号、命之座「金鳍座」、生命值 65,280、水冰抗性 -35%、治疗加成 +75.4% 等极具代入感的战斗面板。

---

## 5. 实现阶段与任务分解

1. **阶段 1：AI 绘画与专属艺术资产生成**
   - 调度绘画 Agent 生成 1 张超清立绘（`splash_vodyanitsa.png`）与 6 张 Q版微表情贴纸（`vodyanitsa_sing.png`、`vodyanitsa_pout.png`、`vodyanitsa_horn.png`、`vodyanitsa_dive.png`、`vodyanitsa_tea.png`、`vodyanitsa_bow.png`）。
   - 存放于 `clients/desktop-web/src/assets/vodyanitsa/images/`。

2. **阶段 2：色盘注册与 CSS 舞台质感**
   - 在 `stores/theme.ts` 中注册 `vodyanitsa` 调色盘；
   - 在 `style.css` 中写入全套 CSS 变量、水波星空背景图案、`.vodyanitsa-dialogue-plate` 材质、凯尔特水妖包角与晶石水滴分割线。

3. **阶段 3：聊天主舞台与气泡体系改造**
   - 在 `ChatStage.vue` 中集成沃雅妮莎对话框渲染逻辑与专属头像框；
   - 支持沃雅妮莎专属贴纸的即时渲染；
   - 工具栏新增贴纸触发器。

4. **阶段 4：专属祈愿池与角色档案展柜**
   - 实现或升级 `VodyanitsaWishModal.vue`（或扩展祈愿抽卡系统），支持沃雅妮莎专属 UP 池；
   - 在 `UserProfileModal.vue` 中为沃雅妮莎定制专属金鳍座展柜。

5. **阶段 5：验证与视觉验收**
   - 构建 `clients/desktop-web` 确保无 TypeScript / Vue 报错；
   - 使用 Chrome DevTools MCP 在本地浏览器中测试并进行截图质检验收；
   - 验证浅色模式与深色模式均保持 100% 清晰高对比度。
