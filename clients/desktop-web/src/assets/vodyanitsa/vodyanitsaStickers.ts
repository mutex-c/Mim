/**
 * 沃雅妮莎（Vodyanitsa）专属微表情贴纸与艺术资产数据源
 * 至冬科洛列夫茨基剧团首席女高音 · 离群的水妖
 */

import splashImg from './images/splash_vodyanitsa.png';
import cheerImg from './images/vodyanitsa_cheer.png';
import confusedImg from './images/vodyanitsa_confused.png';
import cryImg from './images/vodyanitsa_cry.png';
import poutImg from './images/vodyanitsa_pout.png';
import awesomeImg from './images/vodyanitsa_awesome.png';
import sleepImg from './images/vodyanitsa_sleep.png';
import singImg from './images/vodyanitsa_sing.png';
import teaImg from './images/vodyanitsa_tea.png';
import hornImg from './images/vodyanitsa_horn.png';
import diveImg from './images/vodyanitsa_dive.png';
import bowImg from './images/vodyanitsa_bow.png';

export interface VodyanitsaSticker {
  id: string;
  name: string;
  character: string;
  caption: string;
  imageSrc: string;
  svgDataUri: string;
}

export const VODYANITSA_SPLASH_ART = splashImg;

export const VODYANITSA_STICKERS: VodyanitsaSticker[] = [
  {
    id: 'vodyanitsa_cheer',
    name: '沃雅妮莎·好耶！',
    character: '沃雅妮莎',
    caption: '「好耶！今天演出大成功！」',
    imageSrc: cheerImg,
    svgDataUri: cheerImg,
  },
  {
    id: 'vodyanitsa_confused',
    name: '沃雅妮莎·哈？',
    character: '沃雅妮莎',
    caption: '「哈？你在跟本首席开玩笑吗？」',
    imageSrc: confusedImg,
    svgDataUri: confusedImg,
  },
  {
    id: 'vodyanitsa_cry',
    name: '沃雅妮莎·呜呜',
    character: '沃雅妮莎',
    caption: '「呜呜……别再让我背咏叹调乐谱了……」',
    imageSrc: cryImg,
    svgDataUri: cryImg,
  },
  {
    id: 'vodyanitsa_pout',
    name: '沃雅妮莎·哼！',
    character: '沃雅妮莎',
    caption: '「哼！本首席才没有生气呢！」',
    imageSrc: poutImg,
    svgDataUri: poutImg,
  },
  {
    id: 'vodyanitsa_awesome',
    name: '沃雅妮莎·太棒了！',
    character: '沃雅妮莎',
    caption: '「太棒了！今晚的掌声献给你！」',
    imageSrc: awesomeImg,
    svgDataUri: awesomeImg,
  },
  {
    id: 'vodyanitsa_sleep',
    name: '沃雅妮莎·困了',
    character: '沃雅妮莎',
    caption: '「困了……水妖也是需要冬眠的嘛……」',
    imageSrc: sleepImg,
    svgDataUri: sleepImg,
  },
  {
    id: 'vodyanitsa_sing',
    name: '沃雅妮莎·高歌咏叹',
    character: '沃雅妮莎',
    caption: '「水波与音符，在此刻共鸣……」',
    imageSrc: singImg,
    svgDataUri: singImg,
  },
  {
    id: 'vodyanitsa_tea',
    name: '沃雅妮莎·剧院红茶',
    character: '沃雅妮莎',
    caption: '「演出前，一杯热红茶是不可或缺的。」',
    imageSrc: teaImg,
    svgDataUri: teaImg,
  },
];

export const getVodyanitsaStickerById = (id: string): VodyanitsaSticker | undefined => {
  return VODYANITSA_STICKERS.find((s) => s.id === id);
};
