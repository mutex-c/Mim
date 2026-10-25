/**
 * "派蒙的画作" (Paimon's Paintings) - Genshin Impact Chibi Illustrated Stickers
 * Authentic game-style illustrated stickers generated with official character aesthetic.
 */

import paimonEatImg from './images/paimon_eat.png';
import paimonShockImg from './images/paimon_shock.png';
import zhongliTeaImg from './images/zhongli_tea.png';
import hutaoPrankImg from './images/hutao_prank.png';
import furinaCryImg from './images/furina_cry.png';
import raidenMilkImg from './images/raiden_milk.png';
import kleeBombImg from './images/klee_bomb.png';
import nahidaSwingImg from './images/nahida_swing.png';
import ventiEheImg from './images/venti_ehe.png';

export interface ChibiSticker {
  id: string;
  name: string;
  character: string;
  category: 'paimon' | 'archon' | 'liyue' | 'fontaine' | 'mondstadt';
  imageSrc: string;
  svgDataUri: string; // backwards compatibility alias
  caption: string;
}

export const CHIBI_STICKERS: ChibiSticker[] = [
  // 1. 派蒙 - 嗷呜开饭啦
  {
    id: 'chibi_paimon_eat',
    name: '派蒙·大快朵颐',
    character: '派蒙',
    category: 'paimon',
    caption: '好吃！派蒙还要再吃三碗！',
    imageSrc: paimonEatImg,
    svgDataUri: paimonEatImg,
  },

  // 2. 派蒙 - 震惊 / 应急食品？！
  {
    id: 'chibi_paimon_shock',
    name: '派蒙·大吃一惊',
    character: '派蒙',
    category: 'paimon',
    caption: '怎么会这样！你才是应急食品呢！',
    imageSrc: paimonShockImg,
    svgDataUri: paimonShockImg,
  },

  // 3. 钟离 - 优雅品茗 (以普遍理性而论)
  {
    id: 'chibi_zhongli_tea',
    name: '钟离·优雅品茗',
    character: '钟离',
    category: 'liyue',
    caption: '欲买桂花同载酒……只可惜故人，何日再见呢。',
    imageSrc: zhongliTeaImg,
    svgDataUri: zhongliTeaImg,
  },

  // 4. 胡桃 - 嗷呜幽灵 (梅花瞳恶作剧)
  {
    id: 'chibi_hutao_prank',
    name: '胡桃·嗷呜幽灵',
    character: '胡桃',
    category: 'liyue',
    caption: '大丘丘病了二丘丘瞧，三丘丘采药四丘丘嗷呜！',
    imageSrc: hutaoPrankImg,
    svgDataUri: hutaoPrankImg,
  },

  // 5. 芙宁娜 - 戏剧性落泪 (大裁判官救我)
  {
    id: 'chibi_furina_cry',
    name: '芙宁娜·委屈落泪',
    character: '芙宁娜',
    category: 'fontaine',
    caption: '怎么这样！本女士才没有偷吃小蛋糕呢！呜呜呜……',
    imageSrc: furinaCryImg,
    svgDataUri: furinaCryImg,
  },

  // 6. 雷电将军 - 团子牛奶 (甚合我意)
  {
    id: 'chibi_raiden_milk',
    name: '雷电将军·团子牛奶',
    character: '雷电将军',
    category: 'archon',
    caption: '美味，甜度甚佳……此物，亦可纳入永恒。',
    imageSrc: raidenMilkImg,
    svgDataUri: raidenMilkImg,
  },

  // 7. 可莉 - 哒哒哒 (蹦蹦炸弹)
  {
    id: 'chibi_klee_bomb',
    name: '可莉·哒哒哒',
    character: '可莉',
    category: 'mondstadt',
    caption: '可莉又做出新的炸弹啦！琴团长不在吧？哒哒哒~',
    imageSrc: kleeBombImg,
    svgDataUri: kleeBombImg,
  },

  // 8. 纳西妲 - 智慧秋千
  {
    id: 'chibi_nahida_swing',
    name: '纳西妲·智慧秋千',
    character: '纳西妲',
    category: 'archon',
    caption: '最初的贤者啊，请不要忧伤，整个提瓦特都在注视着你。',
    imageSrc: nahidaSwingImg,
    svgDataUri: nahidaSwingImg,
  },

  // 9. 温迪 - 诶嘿 (高天之歌)
  {
    id: 'chibi_venti_ehe',
    name: '温迪·诶嘿',
    character: '温迪',
    category: 'mondstadt',
    caption: '诶嘿~若你困于无风之地，我便为你吹起高天之风。',
    imageSrc: ventiEheImg,
    svgDataUri: ventiEheImg,
  },
];
