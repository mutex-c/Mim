/**
 * 沃雅妮莎（Vodyanitsa）专属微表情贴纸与艺术资产数据源
 * 至冬科洛列夫茨基剧团首席女高音 · 离群的水妖
 */

import splashImg from './images/splash_vodyanitsa.png';
import singImg from './images/vodyanitsa_sing.png';
import poutImg from './images/vodyanitsa_pout.png';
import hornImg from './images/vodyanitsa_horn.png';
import diveImg from './images/vodyanitsa_dive.png';
import teaImg from './images/vodyanitsa_tea.png';
import bowImg from './images/vodyanitsa_bow.png';

export interface VodyanitsaSticker {
  id: string;
  name: string;
  character: string;
  caption: string;
  imageSrc: string;
  svgDataUri: string; // for backwards compatibility with image components
}

export const VODYANITSA_SPLASH_ART = splashImg;

export const VODYANITSA_STICKERS: VodyanitsaSticker[] = [
  {
    id: 'vodyanitsa_sing',
    name: '沃雅妮莎·高歌咏叹',
    character: '沃雅妮莎',
    caption: '「水波与音符，在此刻共鸣……」',
    imageSrc: singImg,
    svgDataUri: singImg,
  },
  {
    id: 'vodyanitsa_pout',
    name: '沃雅妮莎·委屈微嗔',
    character: '沃雅妮莎',
    caption: '「本首席才没有偷懒，只是在润嗓子！」',
    imageSrc: poutImg,
    svgDataUri: poutImg,
  },
  {
    id: 'vodyanitsa_horn',
    name: '沃雅妮莎·唤春角笛',
    character: '沃雅妮莎',
    caption: '「角笛声响，至冬的冰霜也将消融！」',
    imageSrc: hornImg,
    svgDataUri: hornImg,
  },
  {
    id: 'vodyanitsa_dive',
    name: '沃雅妮莎·水妖潜游',
    character: '沃雅妮莎',
    caption: '「水下的世界可比陆地安静多了~」',
    imageSrc: diveImg,
    svgDataUri: diveImg,
  },
  {
    id: 'vodyanitsa_tea',
    name: '沃雅妮莎·剧院红茶',
    character: '沃雅妮莎',
    caption: '「演出前，一杯热红茶是不可或缺的。」',
    imageSrc: teaImg,
    svgDataUri: teaImg,
  },
  {
    id: 'vodyanitsa_bow',
    name: '沃雅妮莎·华丽谢幕',
    character: '沃雅妮莎',
    caption: '「感谢诸位观众，献给今夜最美的回响。」',
    imageSrc: bowImg,
    svgDataUri: bowImg,
  },
];

export const getVodyanitsaStickerById = (id: string): VodyanitsaSticker | undefined => {
  return VODYANITSA_STICKERS.find((s) => s.id === id);
};
