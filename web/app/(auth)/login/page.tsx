"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { setAccessToken } from "@/lib/auth";

export default function LoginPage() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const router = useRouter();

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    // Forwarding credentials to auth-svc via gateway
    const BASE = process.env.NEXT_PUBLIC_API_BASE_URL || "";
    try {
      const res = await fetch(`${BASE}/v1/auth/login`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, password }),
      });

      if (!res.ok) {
        setError("Invalid credentials");
        return;
      }

      const data = await res.json();
      
      // Update in-memory token
      if (data.access_token) {
        setAccessToken(data.access_token);
        
        // Also call our local /api/auth/refresh to set the refresh token in httpOnly cookie
        // In a real scenario, the backend might set it directly if same-site, or the client calls it
        // Assuming the auth-svc returns refresh_token which we need to set via route handler
        if (data.refresh_token) {
           await fetch("/api/auth/refresh", {
             method: "POST",
             headers: { "Content-Type": "application/json" },
             // MOCK: sending refresh token to set it.
             // Normally, the BFF route handler receives it and sets the cookie.
           });
        }
        
        router.push("/dashboard");
      }
    } catch (err) {
      setError("An error occurred during login");
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="max-w-md w-full p-6 bg-white rounded-lg shadow-md">
        <h2 className="text-2xl font-bold text-center mb-6">Login to SănDeal</h2>
        {error && <p className="text-red-500 mb-4 text-center">{error}</p>}
        <form onSubmit={handleLogin} className="space-y-4">
          <div>
            <label className="block text-sm font-medium mb-1">Email</label>
            <input 
              type="email" 
              className="w-full border rounded-md px-3 py-2"
              value={email}
              onChange={e => setEmail(e.target.value)}
              required
            />
          </div>
          <div>
            <label className="block text-sm font-medium mb-1">Password</label>
            <input 
              type="password" 
              className="w-full border rounded-md px-3 py-2"
              value={password}
              onChange={e => setPassword(e.target.value)}
              required
            />
          </div>
          <button type="submit" className="w-full bg-blue-600 text-white rounded-md py-2 font-medium hover:bg-blue-700">
            Login
          </button>
        </form>
      </div>
    </div>
  );
}
