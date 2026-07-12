import styles from "./RatingCircle.module.css";

interface RatingCircleProps {
  value: number;
  size?: number;
}

const RADIUS = 28;
const CIRCUMFERENCE = 2 * Math.PI * RADIUS;

const getColor = (value: number): string => {
  if (value >= 7) return "#1D9E75";
  if (value >= 5) return "#EF9F27";
  return "#E24B4A";
};

const RatingCircle = ({ value, size = 60 }: RatingCircleProps) => {
  const offset = CIRCUMFERENCE - (value / 10) * CIRCUMFERENCE;
  const color = getColor(value);
  const percentage = value * 10;

  return (
    <svg width={size} height={size} viewBox="0 0 60 60" className={styles.svg}>
      <circle cx="30" cy="30" r={RADIUS} fill="none" stroke="var(--color-border-tertiary)" strokeWidth="4" />
      <circle
        cx="30"
        cy="30"
        r={RADIUS}
        fill="none"
        stroke={color}
        strokeWidth="5"
        strokeDasharray={CIRCUMFERENCE}
        strokeDashoffset={offset}
        strokeLinecap="round"
        transform="rotate(-90 30 30)"
      />
      <text x="26" y="36" textAnchor="middle" fontSize="18" fontWeight="500" fill="white">
        {percentage.toFixed(0)}
      </text>
      <text x="41" y="31" textAnchor="middle" fontSize="10" fontWeight="500" fill="white">
        %
      </text>
    </svg>
  );
};

export default RatingCircle;
