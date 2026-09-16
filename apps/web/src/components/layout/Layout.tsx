import Link from "next/link"
import {getSession} from "@/lib/auth/session"
import AuthSessionListener from "./AuthSessionListener"
import LoginLink from "./LoginLink"
import LogoutButton from "./LogoutButton"
import HeaderSearch from "./HeaderSearch"
import RouteLoadingBar from "./RouteLoadingBar"
import styles from "./Layout.module.css"
import TelegramButton from "./TelegramButton"

const Layout = async ({children}: { children: React.ReactNode }) => {
    const session = await getSession()

    return (
        <div className={styles.wrapper}>
            <AuthSessionListener/>
            <header className={styles.header}>
                <div className={styles.headerInner}>
                    <div className={styles.navLeft}>
                        <Link href="/" className={styles.logo}>
                            <img src="/favicon.ico" alt="Movie Tracker" className={styles.icon}/>
                            <span>Movie Tracker</span>
                        </Link>
                        <nav className={styles.nav}>
                            <Link href="/movie" className={styles.navLink}>
                                Фильмы
                            </Link>
                            <Link href="/tv" className={styles.navLink}>
                                Сериалы
                            </Link>
                        </nav>
                    </div>

                    <div className={styles.searchSlot}>
                        <HeaderSearch/>
                    </div>

                    <nav className={`${styles.nav} ${styles.accountNav}`}>
                        <Link href="/watchlist" className={styles.navLink}>
                            Списки
                        </Link>
                        {session.isAuthenticated ? (
                            <>
                                {session.roles.some((role) => role.toLowerCase() === "admin") && (
                                    <Link href="/health" className={styles.navLink}>
                                        Состояние
                                    </Link>
                                )}
                                <div className={styles.logoutContainer}>
                                    {session.username && <p className={styles.username}>{session.username}</p>}
                                    <TelegramButton/>
                                    <LogoutButton/>
                                </div>
                            </>
                        ) : (
                            <LoginLink className={styles.navLink}>
                                Войти
                            </LoginLink>
                        )}
                    </nav>
                </div>
                <RouteLoadingBar/>
            </header>

            <main className={styles.main}>{children}</main>

            <footer className={styles.footer}>
                <p>© 2026 Movie Tracker</p>
            </footer>
        </div>
    )
}

export default Layout
