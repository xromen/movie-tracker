import type { NextConfig } from "next";
import { dirname } from "node:path";
import { fileURLToPath } from "node:url";

const appDir = dirname(fileURLToPath(import.meta.url));

const nextConfig: NextConfig = {
  output: "standalone",
  outputFileTracingIncludes: {
    "/*": ["./node_modules/next/headers.js"],
  },
  reactStrictMode: true,
  turbopack: {
    root: appDir,
  },
};

export default nextConfig;
