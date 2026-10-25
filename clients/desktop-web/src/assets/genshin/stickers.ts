import { CHIBI_STICKERS, ChibiSticker } from './chibiStickers';
export { CHIBI_STICKERS, type ChibiSticker };

export interface GenshinSticker {
  id: string;
  category: 'wish' | 'geo' | 'hydro' | 'electro' | 'dendro' | 'anemo' | 'pyro';
  character: string;
  quote: string;
  badge: string;
  color: string;
  textColor: string;
}

export const GENSHIN_CATEGORIES = [
  { id: 'all', nameZh: '全部台词', icon: '✦' },
  { id: 'wish', nameZh: '抽卡与名梗', icon: '★' },
  { id: 'geo', nameZh: '岩之厚重', icon: '◆' },
  { id: 'hydro', nameZh: '水之盛宴', icon: '◇' },
  { id: 'electro', nameZh: '鸣雷永恒', icon: '✦' },
  { id: 'dendro', nameZh: '草木智慧', icon: '✤' },
  { id: 'anemo', nameZh: '高天之风', icon: '❖' },
  { id: 'pyro', nameZh: '炽烈往生', icon: '▲' },
];

export const GENSHIN_STICKERS: GenshinSticker[] = [
  // 抽卡与名梗
  {
    id: 's_wish_1',
    category: 'wish',
    character: '祈愿',
    quote: '金色传说！十连双黄！★★★★★',
    badge: '欧皇降临',
    color: '#eab308',
    textColor: '#854d0e',
  },
  {
    id: 's_wish_2',
    category: 'wish',
    character: '派蒙',
    quote: '原石 +1600！邮件已发送，快去抽卡吧！✦',
    badge: '紧急补给',
    color: '#38bdf8',
    textColor: '#0369a1',
  },
  {
    id: 's_wish_3',
    category: 'wish',
    character: '派蒙',
    quote: '前面的区域，以后再来探索吧！',
    badge: '世界边界',
    color: '#f59e0b',
    textColor: '#78350f',
  },
  {
    id: 's_wish_4',
    category: 'wish',
    character: '派蒙',
    quote: '不是应急食品！是最好的伙伴啦！',
    badge: '伙伴宣言',
    color: '#f43f5e',
    textColor: '#881337',
  },

  // 岩系 (钟离 / 凝光)
  {
    id: 's_geo_1',
    category: 'geo',
    character: '钟离',
    quote: '欲买桂花同载酒，只可惜故人，何日再见呢。',
    badge: '尘世闲游',
    color: '#d97706',
    textColor: '#78350f',
  },
  {
    id: 's_geo_2',
    category: 'geo',
    character: '钟离',
    quote: '天动万象，此乃天理！',
    badge: '天星降世',
    color: '#b45309',
    textColor: '#ffffff',
  },
  {
    id: 's_geo_3',
    category: 'geo',
    character: '钟离',
    quote: '我虽无意逐鹿，却知苍生苦楚。',
    badge: '磐岩不移',
    color: '#92400e',
    textColor: '#fef3c7',
  },

  // 水系 (芙宁娜 / 那维莱特)
  {
    id: 's_hydro_1',
    category: 'hydro',
    character: '芙宁娜',
    quote: '喝彩吧！沉醉吧！属于我的盛大剧目！',
    badge: '万众瞩目',
    color: '#0284c7',
    textColor: '#e0f2fe',
  },
  {
    id: 's_hydro_2',
    category: 'hydro',
    character: '那维莱特',
    quote: '潮水啊，我已归来！',
    badge: '古龙裁决',
    color: '#0369a1',
    textColor: '#f0f9ff',
  },
  {
    id: 's_hydro_3',
    category: 'hydro',
    character: '芙宁娜',
    quote: '我的演出永不落幕，因为凡人需要信仰！',
    badge: '水仙挽歌',
    color: '#38bdf8',
    textColor: '#0c4a6e',
  },

  // 雷系 (雷电将军 / 八重神子)
  {
    id: 's_elec_1',
    category: 'electro',
    character: '雷电将军',
    quote: '此身，当为稻妻之极致！',
    badge: '一心净土',
    color: '#7c3aed',
    textColor: '#faf5ff',
  },
  {
    id: 's_elec_2',
    category: 'electro',
    character: '雷电将军',
    quote: '无念，断绝！无可留恋！',
    badge: '梦想一心',
    color: '#6d28d9',
    textColor: '#ede9fe',
  },
  {
    id: 's_elec_3',
    category: 'electro',
    character: '八重神子',
    quote: '哎呀呀，这可真是有趣呢～',
    badge: '鸣神宫司',
    color: '#ec4899',
    textColor: '#fdf2f8',
  },

  // 草系 (纳西妲 / 艾尔海森)
  {
    id: 's_dend_1',
    category: 'dendro',
    character: '纳西妲',
    quote: '此即，智慧之殿堂！',
    badge: '真识之芽',
    color: '#16a34a',
    textColor: '#f0fdf4',
  },
  {
    id: 's_dend_2',
    category: 'dendro',
    character: '纳西妲',
    quote: '只要心怀希望，世间所有相遇都是奇迹。',
    badge: '净善遐思',
    color: '#22c55e',
    textColor: '#052e16',
  },

  // 风系 (温迪 / 万叶)
  {
    id: 's_anem_1',
    category: 'anemo',
    character: '温迪',
    quote: '若你困于无风之地，我便为你奏响高天之歌。',
    badge: '高天之音',
    color: '#0d9488',
    textColor: '#ccfbf1',
  },
  {
    id: 's_anem_2',
    category: 'anemo',
    character: '枫原万叶',
    quote: '飘摇风雨中，带刀归来人。',
    badge: '千叶一风',
    color: '#f43f5e',
    textColor: '#fff1f2',
  },
  {
    id: 's_anem_3',
    category: 'anemo',
    character: '温迪',
    quote: '起飞咯～！',
    badge: '风神赐福',
    color: '#14b8a6',
    textColor: '#f0fdfa',
  },

  // 火系 (胡桃 / 可莉)
  {
    id: 's_pyro_1',
    category: 'pyro',
    character: '胡桃',
    quote: '大丘丘病了二丘丘瞧，三丘丘采药四丘丘熬～',
    badge: '幽蝶百戏',
    color: '#ea580c',
    textColor: '#fff7ed',
  },
  {
    id: 's_pyro_2',
    category: 'pyro',
    character: '胡桃',
    quote: '往生堂开业大酬宾，第二份半价哦！',
    badge: '堂主推销',
    color: '#c2410c',
    textColor: '#ffedd5',
  },
  {
    id: 's_pyro_3',
    category: 'pyro',
    character: '可莉',
    quote: '全～部～炸～飞！可莉没有被琴团长关禁闭！',
    badge: '火花骑士',
    color: '#ef4444',
    textColor: '#fef2f2',
  },
];
