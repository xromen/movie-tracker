"use client";

import { Children, useCallback, useEffect, useState, type ReactNode } from "react";
import useEmblaCarousel from "embla-carousel-react";
import { ChevronRight, ChevronLeft } from "lucide-react";
import styles from "./carousel.module.css";

interface CarouselProps {
    title?: string
    children: ReactNode
    className?: string
}

const Carousel = ({ children, className = "", title }: CarouselProps) => {
    const [emblaRef, emblaApi] = useEmblaCarousel({ loop: false, dragFree: true });
    const [prevBtnDisabled, prevBtnDisabledSet] = useState(true)
    const [nextBtnDisabled, nextBtnDisabledSet] = useState(false)

    const onSlideChange = useCallback(() => {
        if (!emblaApi) {
            return
        }

        prevBtnDisabledSet(!emblaApi.canScrollPrev())
        nextBtnDisabledSet(!emblaApi.canScrollNext())
    }, [emblaApi])

    useEffect(() => {
        if (!emblaApi) {
            return
        }

        const initialSyncId = window.setTimeout(onSlideChange, 0)
        emblaApi.on("reInit", onSlideChange)
        emblaApi.on("select", onSlideChange)
        emblaApi.on("slidesInView", onSlideChange)

        return () => {
            window.clearTimeout(initialSyncId)
            emblaApi.off("reInit", onSlideChange)
            emblaApi.off("select", onSlideChange)
            emblaApi.off("slidesInView", onSlideChange)
        }
    }, [emblaApi, onSlideChange])

    return (
        <section className={`${styles.embla} ${className}`} aria-label={title}>
            {title && (
                <h2 className={styles.title}>{title}</h2>
            )}
            <div className={styles.embla__viewport} ref={emblaRef}>
                <div className={styles.embla__container}>
                    {Children.map(children, (child) => (
                        <div className={styles.embla__slide}>
                            {child}
                        </div>
                    ))}
                </div>
            </div>

            <div className={styles.buttons}>
                {!prevBtnDisabled && (
                    <button type="button" aria-label="Прокрутить назад" onClick={() => emblaApi?.scrollPrev()}>
                        <ChevronLeft size={40} aria-hidden="true" />
                    </button>
                )}

                {!nextBtnDisabled && (
                    <button type="button" aria-label="Прокрутить вперед" className={styles.embla__button_next} onClick={() => emblaApi?.scrollNext()}>
                        <ChevronRight size={40} aria-hidden="true" />
                    </button>
                )}
            </div>
        </section>
    );
};

export default Carousel;