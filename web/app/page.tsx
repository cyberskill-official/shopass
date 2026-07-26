import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import { LandingPage } from "@/components/landing/landing-page";
import { refreshCookieCandidates } from "@/lib/server-auth";

export default async function Home() {
  const jar = await cookies();
  const signedIn = refreshCookieCandidates().some((name) => Boolean(jar.get(name)?.value));
  if (signedIn) {
    redirect("/dashboard");
  }
  return <LandingPage />;
}
