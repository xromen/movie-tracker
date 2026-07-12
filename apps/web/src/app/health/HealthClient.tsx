"use client";

import { useEffect, useState, useTransition } from "react";
import { getHealthOverview } from "@/lib/api/health";
import type { DependencyStatus, HealthStatus, PostgresStatus, ReadyResponse } from "@/lib/api/health";
import styles from "./HealthPage.module.css";

const StatusBadge = ({ status }: { status: HealthStatus }) => (
  <span className={`${styles.badge} ${styles[status]}`}>{status}</span>
);

const Value = ({ children }: { children?: React.ReactNode }) => <span className={styles.value}>{children ?? "-"}</span>;

const DependencyCard = ({
  name,
  dependency,
  children,
}: {
  name: string;
  dependency?: DependencyStatus;
  children?: React.ReactNode;
}) => (
  <section className={styles.dependencyCard}>
    <div className={styles.cardHeader}>
      <h2>{name}</h2>
      {dependency ? <StatusBadge status={dependency.status} /> : <span className={styles.unknown}>нет данных</span>}
    </div>
    <dl className={styles.metrics}>
      <div>
        <dt>Задержка</dt>
        <dd>
          <Value>{dependency?.latency}</Value>
        </dd>
      </div>
      {children}
    </dl>
    {dependency?.error && <p className={styles.error}>{dependency.error}</p>}
  </section>
);

const PostgresMetrics = ({ postgres }: { postgres?: PostgresStatus }) => (
  <>
    <div>
      <dt>Всего соединений</dt>
      <dd>
        <Value>{postgres?.total_conns}</Value>
      </dd>
    </div>
    <div>
      <dt>Занято соединений</dt>
      <dd>
        <Value>{postgres?.acquired_conns}</Value>
      </dd>
    </div>
    <div>
      <dt>Свободно соединений</dt>
      <dd>
        <Value>{postgres?.idle_conns}</Value>
      </dd>
    </div>
    <div>
      <dt>Ожиданий соединения</dt>
      <dd>
        <Value>{postgres?.empty_acquire_count}</Value>
      </dd>
    </div>
    <div>
      <dt>Суммарное ожидание</dt>
      <dd>
        <Value>{postgres?.empty_acquire_wait_time}</Value>
      </dd>
    </div>
  </>
);

const HealthClient = () => {
  const [liveOk, setLiveOk] = useState(false);
  const [ready, setReady] = useState<ReadyResponse>();
  const [isError, setIsError] = useState(false);
  const [isPending, startTransition] = useTransition();

  const refresh = () => {
    startTransition(async () => {
      setIsError(false);

      try {
        const overview = await getHealthOverview();

        setLiveOk(overview.liveOk);
        setReady(overview.ready);

        if (!overview.liveOk || !overview.ready) {
          setIsError(true);
        }
      } catch {
        setLiveOk(false);
        setReady(undefined);
        setIsError(true);
      }
    });
  };

  useEffect(() => {
    refresh();
    const intervalId = window.setInterval(refresh, 30_000);

    return () => window.clearInterval(intervalId);
  }, []);

  return (
    <div className={styles.page}>
      <div className={styles.topbar}>
        <div>
          <p className={styles.eyebrow}>MONITORING</p>
          <h1>Состояние приложения</h1>
          <p className={styles.description}>Проверка обновляется автоматически каждые 30 секунд.</p>
        </div>
        <button className={styles.refreshButton} onClick={refresh} disabled={isPending}>
          {isPending ? "Обновление..." : "Обновить"}
        </button>
      </div>

      <section className={styles.summary}>
        <div className={styles.summaryItem}>
          <span>Приложение запущено</span>
          {isPending ? <span className={styles.unknown}>проверка...</span> : liveOk ? <StatusBadge status="ok" /> : <StatusBadge status="unavailable" />}
        </div>
        <div className={styles.summaryItem}>
          <span>Готовность к работе</span>
          {isPending ? (
            <span className={styles.unknown}>проверка...</span>
          ) : ready ? (
            <StatusBadge status={ready.status} />
          ) : (
            <StatusBadge status="unavailable" />
          )}
        </div>
      </section>

      {isError && (
        <p className={styles.requestError}>
          Не удалось получить часть данных о состоянии. Проверьте доступность API и повторите попытку.
        </p>
      )}

      <section className={styles.applicationCard}>
        <div className={styles.cardHeader}>
          <h2>Приложение</h2>
          {ready ? <StatusBadge status={ready.status} /> : <span className={styles.unknown}>нет данных</span>}
        </div>
        <dl className={styles.metrics}>
          <div>
            <dt>Версия</dt>
            <dd>
              <Value>{ready?.version}</Value>
            </dd>
          </div>
          <div>
            <dt>Время работы</dt>
            <dd>
              <Value>{ready?.uptime}</Value>
            </dd>
          </div>
          <div>
            <dt>Горутины</dt>
            <dd>
              <Value>{ready?.goroutines}</Value>
            </dd>
          </div>
        </dl>
      </section>

      <div className={styles.dependencies}>
        <DependencyCard name="PostgreSQL" dependency={ready?.dependencies.postgres}>
          <PostgresMetrics postgres={ready?.dependencies.postgres} />
        </DependencyCard>
        <DependencyCard name="Redis" dependency={ready?.dependencies.redis} />
      </div>
    </div>
  );
};

export default HealthClient;
