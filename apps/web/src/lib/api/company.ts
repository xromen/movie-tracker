import { fetchApi } from "./client"
import type { CompanyDetails } from "./types"

interface ApiCompanyDetails {
    id: number
    name: string
    description?: string
    logo_path?: string
    origin_country?: string
    homepage?: string
    headquarters?: string
    headquearters?: string
}

export const getCompanyDetails = async (companyId: number): Promise<CompanyDetails> => {
    const company = await fetchApi<ApiCompanyDetails>(`/v1/company/${companyId}`, { revalidate: 3600 })

    return {
        id: company.id,
        name: company.name,
        description: company.description,
        logoPath: company.logo_path,
        originCountry: company.origin_country,
        homepage: company.homepage,
        headquarters: company.headquarters ?? company.headquearters,
    }
}
