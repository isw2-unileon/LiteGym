import { mkdir } from "node:fs/promises";
import path from "node:path";

import { request, type FullConfig } from "@playwright/test";

const users = [
  {
    email: "diego@example.com",
    password: "1234",
    storageFile: "admin.json",
  },
  {
    email: "laura@example.com",
    password: "1234",
    storageFile: "laura.json",
  },
];

export default async function globalSetup(config: FullConfig) {
  const baseURL = config.projects[0]?.use?.baseURL ?? "http://localhost:5173";
  const authDir = path.join(__dirname, ".auth");

  await mkdir(authDir, { recursive: true });

  for (const user of users) {
    const apiContext = await request.newContext({ baseURL });
    const response = await apiContext.post("/api/auth/login", {
      data: {
        email: user.email,
        password: user.password,
      },
    });

    if (!response.ok()) {
      throw new Error(`Failed to create auth state for ${user.email}: ${response.status()}`);
    }

    await apiContext.storageState({
      path: path.join(authDir, user.storageFile),
    });
    await apiContext.dispose();
  }
}
