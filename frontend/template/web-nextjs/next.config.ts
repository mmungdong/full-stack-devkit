import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // Static export: generate pure static HTML/CSS/JS
  output: "export",

  // Disable image optimization (not compatible with static export)
  images: {
    unoptimized: true,
  },

  // Trailing slashes for static hosting compatibility
  trailingSlash: true,

  // Transpile @devkit/shared since it ships as TypeScript source
  transpilePackages: ["@devkit/shared"],
};

export default nextConfig;
