import React from 'react'
import { DocsThemeConfig } from 'nextra-theme-docs'
import Image  from 'next/image';
import logo from './logo.png';

const config: DocsThemeConfig = {
  logo: <Image src={logo} width={48} alt="TofuTF Logo" />,
  primaryHue: 231,
  useNextSeoProps:()=>({
    titleTemplate: '%s | TofuTF'
  }),
  project: {
    link: 'https://github.com/tofutf/tofutf',
  },
  docsRepositoryBase: 'https://github.com/tofutf/tofutf-documentation',
  footer: {
    text: (
      <span>
        MIT {new Date().getFullYear()} ©{' '}
        <a href="https://github.com/tofutf/tofutf" target="_blank">
          TofuTF
        </a>
        .
      </span>
    ),
  },
  banner: {
    key: 'under-construction',
    text: (
      <p>🏗️🚧 The <b>TofuTF</b> docs are under construction. </p>
    )
  }
}

export default config
