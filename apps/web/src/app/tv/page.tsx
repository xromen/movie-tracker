import type { Metadata } from "next";
import TvCatalog from "./TvCatalog";

interface TvPageProps {
  searchParams: Promise<{
    filter?: string;
    page?: string;
  }>;
}

export const metadata: Metadata = {
  title: "Сериалы - Movie Tracker",
  description: "Каталог сериалов Movie Tracker: поиск, популярные сериалы, рейтинги, сезоны и рекомендации.",
  alternates: {
    canonical: "/tv",
  },
};

const TvPage = async ({ searchParams }: TvPageProps) => {
  const params = await searchParams;
  const page = Math.max(1, Number(params.page ?? 1) || 1);

  return <TvCatalog filter={params.filter} page={page} />;
};

export default TvPage;
