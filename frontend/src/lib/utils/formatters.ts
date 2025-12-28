export const formatDate = (date: string): string => {
	return new Date(date).toLocaleString();
};

export const truncateText = (text: string, maxLength: number): string => {
	if (text.length <= maxLength) return text;
	return `${text.substring(0, maxLength)}...`;
};

export const daysSince = (date: string): number => {
	const then = new Date(date);
	const now = new Date();
	const diffMs = now.getTime() - then.getTime();
	const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));
	return diffDays;
};

export interface SLAStatus {
	daysRemaining: number;
	status: 'compliant' | 'warning' | 'exceeded';
	color: string;
	bgColor: string;
}

export const calculateSLAStatus = (firstDetectedAt: string, severity: string, slaLimits: { critical: number; high: number; medium: number; low: number }): SLAStatus => {
	const daysElapsed = daysSince(firstDetectedAt);

	let slaLimit: number;
	switch (severity) {
		case 'Critical':
			slaLimit = slaLimits.critical;
			break;
		case 'High':
			slaLimit = slaLimits.high;
			break;
		case 'Medium':
			slaLimit = slaLimits.medium;
			break;
		case 'Low':
			slaLimit = slaLimits.low;
			break;
		default:
			slaLimit = 180; // Default for unknown severity
	}

	const daysRemaining = slaLimit - daysElapsed;

	let status: 'compliant' | 'warning' | 'exceeded';
	let color: string;
	let bgColor: string;

	if (daysRemaining < 0) {
		status = 'exceeded';
		color = 'text-red-700';
		bgColor = 'bg-red-50';
	} else if (daysRemaining <= Math.floor(slaLimit * 0.2)) {
		// Warning when less than 20% of SLA time remains
		status = 'warning';
		color = 'text-yellow-700';
		bgColor = 'bg-yellow-50';
	} else {
		status = 'compliant';
		color = 'text-green-700';
		bgColor = 'bg-green-50';
	}

	return { daysRemaining, status, color, bgColor };
};
