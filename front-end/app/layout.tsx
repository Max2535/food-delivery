import type { Metadata } from "next";
import Providers from "./providers";
import Navbar from "./navbar";
import "./globals.css";

export const metadata: Metadata = {
  title: "Food Delivery",
  description: "Food Delivery Application",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="th">
      <body className="bg-gray-50 min-h-screen">
        <Providers>
          <Navbar />
          <main>{children}</main>
        </Providers>
      </body>
    </html>
  );
}
