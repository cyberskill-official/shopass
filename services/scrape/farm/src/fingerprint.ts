export interface FingerprintProfile {
  userAgent: string;
  platform: string;            // 'Win32' | 'Linux x86_64' | ...
  languages: string[];         // ['vi-VN','vi','en-US']
  timezone: string;            // 'Asia/Ho_Chi_Minh'
  screen: { width: number; height: number; dpr: number };
  hardwareConcurrency: number;
  deviceMemory: number;
  webgl: { vendor: string; renderer: string };
  canvasNoiseSeed: number;     // seed ổn định cho readback spoof của profile
}

function pseudoRandom(seed: number) {
  let x = Math.sin(seed++) * 10000;
  return x - Math.floor(x);
}

export function makeProfile(country: 'VN' | 'ID' | 'TH', seed: number): FingerprintProfile {
  const isMobile = pseudoRandom(seed) > 0.5;
  const isWin = pseudoRandom(seed + 1) > 0.3;
  
  let timezone = 'Asia/Ho_Chi_Minh';
  let languages = ['vi-VN', 'vi', 'en-US'];
  
  if (country === 'ID') {
    timezone = 'Asia/Jakarta';
    languages = ['id-ID', 'id', 'en-US'];
  } else if (country === 'TH') {
    timezone = 'Asia/Bangkok';
    languages = ['th-TH', 'th', 'en-US'];
  }

  // Consistent UA and Platform
  let userAgent = 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36';
  let platform = 'Win32';
  let webgl = { vendor: 'Google Inc. (NVIDIA)', renderer: 'ANGLE (NVIDIA, NVIDIA GeForce RTX 3060 Direct3D11 vs_5_0 ps_5_0, D3D11)' };

  if (!isWin) {
    platform = 'MacIntel';
    userAgent = 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36';
    webgl = { vendor: 'Google Inc. (Apple)', renderer: 'ANGLE (Apple, Apple M1, OpenGL 4.1)' };
  }

  return {
    userAgent,
    platform,
    languages,
    timezone,
    screen: { width: 1920, height: 1080, dpr: 1 },
    hardwareConcurrency: 8,
    deviceMemory: 8,
    webgl,
    canvasNoiseSeed: seed,
  };
}

export function isCoherent(p: FingerprintProfile): boolean {
  // timezone <-> locale coherence
  if (p.timezone === 'Asia/Ho_Chi_Minh' && !p.languages.includes('vi-VN')) return false;
  if (p.timezone === 'Asia/Jakarta' && !p.languages.includes('id-ID')) return false;
  if (p.timezone === 'Asia/Bangkok' && !p.languages.includes('th-TH')) return false;

  // platform <-> UA coherence
  if (p.platform === 'Win32' && !p.userAgent.includes('Windows')) return false;
  if (p.platform === 'MacIntel' && !p.userAgent.includes('Macintosh')) return false;

  return true;
}
