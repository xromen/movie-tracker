import {MediaType} from "@/lib/api/types";
import MediaDetails from "@/app/details/[mediaType]/[mediaId]/MediaDetails";


interface DetailsPageProps {
    params: Promise<{
        mediaType: MediaType;
        mediaId: number;
    }>;
}

const DetailsPage = async ({params}: DetailsPageProps) => {
    const {mediaType, mediaId} = await params;

    return (
        <MediaDetails mediaType={mediaType} mediaId={mediaId}></MediaDetails>
    )
};

export default DetailsPage;
