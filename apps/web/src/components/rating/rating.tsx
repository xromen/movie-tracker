import styles from "./rating.module.css"

export interface RatingProps {
    voteAverage?: number
    voteCount?: number
}

const getColor = (value?: number): string => {
    if (!value) return "";
    if (value >= 7) return "#1D9E75";
    if (value >= 5) return "#EF9F27";
    return "#E24B4A";
};

const Rating = ({ voteAverage }: RatingProps) => {
    if (!voteAverage || voteAverage === 0) {
        return null
    }

    const ratingColor = getColor(voteAverage)

    return (
        <span className={styles.wrapper} style={{ background: ratingColor }}>
            {voteAverage.toFixed(1)}
        </span>
    )
}

export default Rating