import { redirect } from "@tanstack/react-router";
import { getStoredUser } from "@/lib/auth";

/**
 * Checks if the current user is an instance admin by examining the stored user's role.
 * Returns true if the user has the 'instance_admin' role.
 */
export function checkIsInstanceAdmin(): boolean {
  const user = getStoredUser();
  if (!user) return false;
  if ("role" in user && user.role) {
    return Array.isArray(user.role)
      ? user.role.includes("instance_admin")
      : user.role === "instance_admin";
  }
  return false;
}

/**
 * beforeLoad guard for instance-only routes.
 * Redirects to /403 if the user is not an instance admin.
 * Must be used inside a beforeLoad callback where redirect() can be thrown.
 */
export function requireInstanceAdmin(): void {
  if (!checkIsInstanceAdmin()) {
    throw redirect({ to: "/403" });
  }
}
