import Link from "next/link";
import styles from "./Pagination.module.css";

interface PaginationProps {
  currentPage: number;
  totalPages: number;
  pathname: string;
  searchParams?: Record<string, string | number | undefined>;
  pageParam?: string;
}

const getPageHref = (
  pathname: string,
  searchParams: Record<string, string | number | undefined>,
  pageParam: string,
  page: number,
) => {
  const params = new URLSearchParams();

  Object.entries(searchParams).forEach(([key, value]) => {
    if (value !== undefined && value !== "") {
      params.set(key, String(value));
    }
  });

  if (page > 1) {
    params.set(pageParam, String(page));
  } else {
    params.delete(pageParam);
  }

  const query = params.toString();

  return query ? `${pathname}?${query}` : pathname;
};

const pageButtonsCount = 7

const getPages = (currentPage: number, totalPages: number) => {
  const visiblePages = Math.min(pageButtonsCount, totalPages);

  if (totalPages <= pageButtonsCount) {
    return Array.from({ length: totalPages }, (_, index) => index + 1);
  }

  const half = Math.floor(visiblePages / 2);

  let start = currentPage - half;
  let end = currentPage + half;

  if (visiblePages % 2 === 0) {
    end -= 1;
  }

  if (start <= 2) {
    start = 1;
    end = visiblePages;
  }

  if (end >= totalPages - 1) {
    end = totalPages;
    start = totalPages - visiblePages + 1;
  }

  return Array.from({ length: end - start + 1 }, (_, index) => start + index);
};

const Pagination = ({
  currentPage,
  totalPages,
  pathname,
  searchParams = {},
  pageParam = "page",
}: PaginationProps) => {
  if (totalPages <= 1) {
    return null;
  }

  const normalizedPage = Math.min(Math.max(currentPage, 1), totalPages);
  const pages = getPages(normalizedPage, totalPages);

  return (
    <nav className={styles.pagination} aria-label="Пагинация">
      {!pages.includes(1) && (
        <>
          <Link key={totalPages} className={styles.link} href={getPageHref(pathname, searchParams, pageParam, 1)}>
            1
          </Link>
          <span>...</span>
        </>
      )}

      {pages.map((page) =>
        page === normalizedPage ? (
          <span key={page} className={styles.current} aria-current="page">
            {page}
          </span>
        ) : (
          <Link key={page} className={styles.link} href={getPageHref(pathname, searchParams, pageParam, page)}>
            {page}
          </Link>
        ),
      )}

      {!pages.includes(totalPages) && (
        <>
          <span>...</span>
          <Link key={totalPages} className={styles.link} href={getPageHref(pathname, searchParams, pageParam, totalPages)}>
            {totalPages}
          </Link>
        </>
      )}
    </nav>
  );
};

export default Pagination;
