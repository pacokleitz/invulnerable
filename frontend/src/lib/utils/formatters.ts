export const formatDate = (date: string): string => {
	return new Date(date).toLocaleString();
};

export const truncateText = (text: string, maxLength: number): string => {
	if (text.length <= maxLength) return text;
	return `${text.substring(0, maxLength)}...`;
};
