/** @type {import('next').NextConfig} */
const nextConfig = {
  experimental: {
    serverActions: true,
  },
  webpack: (config) => {
    config.externals = [...config.externals, '@google-cloud/bigquery'];
    return config;
  },
};

export default nextConfig;
