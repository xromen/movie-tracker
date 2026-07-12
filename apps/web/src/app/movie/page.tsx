import type { Metadata } from "next";
import MovieCatalog from "./MovieCatalog";

interface MoviePageProps {
  searchParams: Promise<{
    filter?: string;
    page?: string;
  }>;
}

export const metadata: Metadata = {
  title: "Фильмы - Movie Tracker",
  description: "Каталог фильмов Movie Tracker: поиск, популярные фильмы, новинки, рейтинги и рекомендации.",
  alternates: {
    canonical: "/movie",
  },
};

const MoviePage = async ({ searchParams }: MoviePageProps) => {
  const params = await searchParams;
  const page = Math.max(1, Number(params.page ?? 1) || 1);

  return <MovieCatalog filter={params.filter} page={page} />;
};

export default MoviePage;
