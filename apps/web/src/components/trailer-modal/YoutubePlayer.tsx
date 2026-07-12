interface YoutubePlayerProps {
  videoKey: string;
  title: string;
}

const YoutubePlayer = ({ videoKey, title = "Trailer" }: YoutubePlayerProps) => {
  return (
    <div style={{ position: "relative", paddingBottom: "56.25%", height: 0 }}>
      <iframe
        style={{
          position: "absolute",
          top: 0,
          left: 0,
          width: "100%",
          height: "100%",
          borderRadius: "8px",
          border: "none",
        }}
        src={`https://www.youtube.com/embed/${videoKey}`}
        title={title}
        allowFullScreen
      />
    </div>
  );
};

export default YoutubePlayer;
