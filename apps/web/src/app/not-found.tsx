import Link from "next/link";

const NotFound = () => {
  return (
    <section style={{ padding: 24 }}>
      <h1>Страница не найдена</h1>
      <p>Такой страницы нет или она была перемещена.</p>
      <Link href="/movie">Вернуться к фильмам</Link>
    </section>
  );
};

export default NotFound;
