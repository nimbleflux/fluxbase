import axios, { type AxiosError, type AxiosInstance } from "axios";
import { useAuthStore } from "@/stores/auth-store";
import { useTenantStore } from "@/stores/tenant-store";
import { useBranchStore } from "@/stores/branch-store";

export const API_BASE_URL =
  window.__FLUXBASE_CONFIG__?.publicBaseURL ||
  import.meta.env.VITE_API_URL ||
  "";

export const api: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    "Content-Type": "application/json",
  },
  timeout: 30000,
});

let isRefreshing = false;
let failedQueue: Array<{
  resolve: (value: unknown) => void;
  reject: (reason?: unknown) => void;
}> = [];

const processQueue = (error: Error | null, token: string | null = null) => {
  failedQueue.forEach((prom) => {
    if (error) {
      prom.reject(error);
    } else {
      prom.resolve(token);
    }
  });
  failedQueue = [];
};

const isNotLoggedInResponse = (data: unknown): boolean => {
  if (!data || typeof data !== "object") return false;
  const obj = data as Record<string, unknown>;
  const errorFields = [obj.error, obj.message, obj.msg, obj.detail];
  for (const field of errorFields) {
    if (typeof field === "string") {
      const lower = field.toLowerCase();
      if (
        lower.includes("not logged in") ||
        lower.includes("not authenticated") ||
        lower.includes("unauthorized") ||
        lower.includes("invalid token") ||
        lower.includes("token expired") ||
        lower.includes("session expired") ||
        lower.includes("authentication required")
      ) {
        return true;
      }
    }
  }
  return false;
};

api.interceptors.request.use(
  (config) => {
    if (!config.headers.Authorization) {
      const { accessToken } = useAuthStore.getState().auth;
      if (accessToken) {
        config.headers.Authorization = `Bearer ${accessToken}`;
      }
    }

    try {
      const currentTenant = useTenantStore.getState().currentTenant;
      if (currentTenant?.id) {
        config.headers["X-FB-Tenant"] = currentTenant.id;
      }
    } catch {
      /* Intentionally empty: tenant store may not be available */
    }

    try {
      const currentBranch = useBranchStore.getState().currentBranch;
      if (currentBranch?.slug && currentBranch.type !== "main") {
        config.headers["X-Fluxbase-Branch"] = currentBranch.slug;
      }
    } catch {
      /* Intentionally empty: branch store may not be available */
    }

    return config;
  },
  (error) => Promise.reject(error),
);

api.interceptors.response.use(
  (response) => {
    if (isNotLoggedInResponse(response.data)) {
      const { refreshToken } = useAuthStore.getState().auth;
      if (refreshToken) {
        return axios
          .post(`${API_BASE_URL}/api/v1/admin/refresh`, {
            refresh_token: refreshToken,
          })
          .then((refreshResponse) => {
            const {
              access_token,
              refresh_token: newRefreshToken,
              user,
              expires_in,
            } = refreshResponse.data;
            const store = useAuthStore.getState().auth;
            store.setTokens(access_token, newRefreshToken);
            if (user) {
              store.setUser({
                accountNo: user.id,
                email: user.email,
                role: [user.role || "tenant_admin"],
                exp: Date.now() + expires_in * 1000,
              });
            }
            if (response.config.headers) {
              response.config.headers.Authorization = `Bearer ${access_token}`;
            }
            return api(response.config);
          })
          .catch(() => {
            useAuthStore.getState().auth.reset();
            window.location.href = "/admin/login";
            return new Promise(() => {});
          });
      }
      useAuthStore.getState().auth.reset();
      window.location.href = "/admin/login";
      return new Promise(() => {});
    }
    return response;
  },
  async (error: AxiosError) => {
    const originalRequest = error.config as typeof error.config & {
      _retry?: boolean;
    };
    const url = originalRequest?.url || "";
    const skipRefreshPaths = ["/api/v1/admin/branches"];
    const shouldSkipRefresh = skipRefreshPaths.some((path) =>
      url.startsWith(path),
    );

    if (
      error.response?.status === 401 &&
      !originalRequest._retry &&
      !shouldSkipRefresh
    ) {
      if (isRefreshing) {
        return new Promise((resolve, reject) => {
          failedQueue.push({ resolve, reject });
        })
          .then((token) => {
            if (originalRequest.headers) {
              originalRequest.headers.Authorization = `Bearer ${token}`;
            }
            return api(originalRequest);
          })
          .catch((err) => Promise.reject(err));
      }

      originalRequest._retry = true;
      isRefreshing = true;

      const refreshToken = useAuthStore.getState().auth.refreshToken;
      if (!refreshToken) {
        useAuthStore.getState().auth.reset();
        window.location.href = "/admin/login";
        return new Promise(() => {});
      }

      try {
        const response = await axios.post(
          `${API_BASE_URL}/api/v1/admin/refresh`,
          { refresh_token: refreshToken },
        );
        const {
          access_token,
          refresh_token: newRefreshToken,
          user,
          expires_in,
        } = response.data;
        const store = useAuthStore.getState().auth;
        store.setTokens(access_token, newRefreshToken);
        if (user) {
          store.setUser({
            accountNo: user.id,
            email: user.email,
            role: [user.role || "tenant_admin"],
            exp: Date.now() + expires_in * 1000,
          });
        }
        if (originalRequest.headers) {
          originalRequest.headers.Authorization = `Bearer ${access_token}`;
        }
        processQueue(null, access_token);
        isRefreshing = false;
        return api(originalRequest);
      } catch (refreshError) {
        processQueue(refreshError as Error, null);
        isRefreshing = false;
        useAuthStore.getState().auth.reset();
        window.location.href = "/admin/login";
        return new Promise(() => {});
      }
    }

    if (
      error.response?.data &&
      isNotLoggedInResponse(error.response.data) &&
      !originalRequest._retry
    ) {
      originalRequest._retry = true;
      const refreshToken = useAuthStore.getState().auth.refreshToken;
      if (refreshToken) {
        try {
          const response = await axios.post(
            `${API_BASE_URL}/api/v1/admin/refresh`,
            { refresh_token: refreshToken },
          );
          const {
            access_token,
            refresh_token: newRefreshToken,
            user,
            expires_in,
          } = response.data;
          const store = useAuthStore.getState().auth;
          store.setTokens(access_token, newRefreshToken);
          if (user) {
            store.setUser({
              accountNo: user.id,
              email: user.email,
              role: [user.role || "tenant_admin"],
              exp: Date.now() + expires_in * 1000,
            });
          }
          if (originalRequest.headers) {
            originalRequest.headers.Authorization = `Bearer ${access_token}`;
          }
          return api(originalRequest);
        } catch {
          useAuthStore.getState().auth.reset();
          window.location.href = "/admin/login";
          return new Promise(() => {});
        }
      }
      useAuthStore.getState().auth.reset();
      window.location.href = "/admin/login";
      return new Promise(() => {});
    }

    return Promise.reject(error);
  },
);

export const getDashboardAuthHeaders = (): Record<string, string> => {
  const { accessToken } = useAuthStore.getState().auth;
  return accessToken ? { Authorization: `Bearer ${accessToken}` } : {};
};
