package rules

import (
	"regexp"
	"strings"

	"ingress-migration-analyzer/internal/models"
)

// GetAnnotationRules returns the complete set of annotation classification rules
func GetAnnotationRules() []models.AnnotationRule {
	return []models.AnnotationRule{
		// Tier A - AUTO (annotations with established Gateway API equivalents)
		{
			Name:        "Rewrite Target",
			Pattern:     "nginx.ingress.kubernetes.io/rewrite-target",
			RiskLevel:   models.RiskAuto,
			Description: "URL path rewriting functionality",
			MigrationNote: "Gateway API HTTPRoute supports path rewriting via URLRewrite filters (GEP-726). " +
				"Most Gateway implementations support this feature.",
			SourceURL: "https://gateway-api.sigs.k8s.io/guides/http-redirect-rewrite/",
		},
		{
			Name:        "SSL Redirect",
			Pattern:     "nginx.ingress.kubernetes.io/ssl-redirect",
			RiskLevel:   models.RiskAuto,
			Description: "Automatic HTTPS redirect",
			MigrationNote: "Gateway API HTTPRoute supports HTTPS redirects via RequestRedirect filters. " +
				"Standard feature across Gateway implementations.",
			SourceURL: "https://gateway-api.sigs.k8s.io/guides/http-redirect-rewrite/",
		},
		{
			Name:        "Force SSL Redirect",
			Pattern:     "nginx.ingress.kubernetes.io/force-ssl-redirect",
			RiskLevel:   models.RiskAuto,
			Description: "Force HTTPS redirect even for non-SSL listeners",
			MigrationNote: "Gateway API HTTPRoute supports HTTPS redirects via RequestRedirect filters. " +
				"Similar implementation pattern to ssl-redirect.",
			SourceURL: "https://gateway-api.sigs.k8s.io/guides/http-redirect-rewrite/",
		},
		{
			Name:        "Backend Protocol",
			Pattern:     "nginx.ingress.kubernetes.io/backend-protocol",
			RiskLevel:   models.RiskManual,
			Description: "Specifies protocol for backend communication (HTTP/HTTPS/GRPC/etc)",
			MigrationNote: "Gateway API BackendRef supports protocol fields, but implementation " +
				"varies by Gateway provider. Verify your Gateway supports the required protocols.",
			SourceURL: "https://gateway-api.sigs.k8s.io/reference/spec/#backendref",
		},
		{
			Name:        "Use Regex",
			Pattern:     "nginx.ingress.kubernetes.io/use-regex",
			RiskLevel:   models.RiskManual,
			Description: "Enable regex matching for paths",
			MigrationNote: "Gateway API HTTPRoute supports RegularExpression path matching (v1.1+). " +
				"Verify your Gateway implementation supports regex and review syntax differences.",
			SourceURL: "https://gateway-api.sigs.k8s.io/reference/spec/#httppathmatch",
		},

		// Tier B - MANUAL (medium complexity, requires review)
		{
			Name:        "Proxy Body Size",
			Pattern:     "nginx.ingress.kubernetes.io/proxy-body-size",
			RiskLevel:   models.RiskManual,
			Description: "Maximum size of the client request body",

			MigrationNote: "No standardized Gateway API equivalent. Gateway implementations may support " +
				"request size limits via vendor-specific policies. Check your Gateway documentation.",
			SourceURL: "https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/annotations/#proxy-body-size",
		},
		{
			Name:        "Proxy Read Timeout",
			Pattern:     "nginx.ingress.kubernetes.io/proxy-read-timeout",
			RiskLevel:   models.RiskManual,
			Description: "Timeout for reading response from backend",
			MigrationNote: "Gateway API may support timeouts via implementation-specific policies. " +
				"Check your Gateway implementation's policy support or use service mesh.",
			SourceURL: "https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/annotations/#proxy-read-timeout",
		},
		{
			Name:        "Proxy Send Timeout",
			Pattern:     "nginx.ingress.kubernetes.io/proxy-send-timeout",
			RiskLevel:   models.RiskManual,
			Description: "Timeout for transmitting request to backend",
			MigrationNote: "Similar to read timeout - check Gateway implementation policy support " +
				"or implement at application/service mesh level.",
			SourceURL: "https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/annotations/#proxy-send-timeout",
		},
		{
			Name:        "Auth URL",
			Pattern:     "nginx.ingress.kubernetes.io/auth-url",
			RiskLevel:   models.RiskManual,
			Description: "External authentication service URL",
			MigrationNote: "Gateway API doesn't standardize external auth, but many implementations " +
				"support it. Consider OAuth2/OIDC policies or service mesh auth instead.",
			SourceURL: "https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/annotations/#auth-url",
		},
		{
			Name:        "Proxy Connect Timeout",
			Pattern:     "nginx.ingress.kubernetes.io/proxy-connect-timeout",
			RiskLevel:   models.RiskManual,
			Description: "Timeout for establishing connection to backend",
			MigrationNote: "Check Gateway implementation support for connection timeouts or " +
				"implement circuit breaker patterns at the application level.",
			SourceURL: "https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/annotations/#proxy-connect-timeout",
		},
		{
			Name:        "Client Body Buffer Size",
			Pattern:     "nginx.ingress.kubernetes.io/client-body-buffer-size",
			RiskLevel:   models.RiskManual,
			Description: "Buffer size for reading client request body",
			MigrationNote: "Implementation-specific setting. Review if your application requires " +
				"specific buffering behavior and implement accordingly.",
			SourceURL: "https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/annotations/#client-body-buffer-size",
		},
		{
			Name:        "CORS Enable",
			Pattern:     "nginx.ingress.kubernetes.io/enable-cors",
			RiskLevel:   models.RiskManual,
			Description: "Enable CORS headers",
			MigrationNote: "Some Gateway implementations support CORS via policies. " +
				"Alternatively, implement CORS at application level or via service mesh.",
			SourceURL: "https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/annotations/#enable-cors",
		},
		{
			Name:        "Rate Limiting",
			Pattern:     "nginx.ingress.kubernetes.io/rate-limit",
			RiskLevel:   models.RiskManual,
			Description: "Request rate limiting configuration",
			MigrationNote: "Gateway API is developing rate limiting standards (GEP-1731). " +
				"Check your Gateway implementation or use service mesh rate limiting.",
			SourceURL: "https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/annotations/#rate-limiting",
		},

		// Tier C - HIGH_RISK (complex configurations needing careful planning)
		{
			Name:        "Server Snippet",
			Pattern:     "nginx.ingress.kubernetes.io/server-snippet",
			RiskLevel:   models.RiskHigh,
			Description: "Custom NGINX server block configuration",
			MigrationNote: "Server snippets contain custom NGINX configuration that has no Gateway API equivalent. " +
				"Review the configuration and implement equivalent functionality using Gateway policies, " +
				"service mesh, or consider staying with NGINX Inc commercial controller.",
			SourceURL: "https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/annotations/#server-snippet",
		},
		{
			Name:        "Configuration Snippet",
			Pattern:     "nginx.ingress.kubernetes.io/configuration-snippet",
			RiskLevel:   models.RiskHigh,
			Description: "Custom NGINX location block configuration",
			MigrationNote: "Configuration snippets require manual analysis and reimplementation. " +
				"Consider Gateway API policies, service mesh capabilities, or application-level changes.",
			SourceURL: "https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/annotations/#configuration-snippet",
		},
		{
			Name:        "Location Snippet",
			Pattern:     "nginx.ingress.kubernetes.io/location-snippet",
			RiskLevel:   models.RiskHigh,
			Description: "Custom NGINX location configuration",
			MigrationNote: "Location snippets need careful review for functionality. " +
				"Map to Gateway API filters, policies, or service mesh configurations where possible.",
			SourceURL: "https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/annotations/#configuration-snippet",
		},
		{
			Name:        "Stream Snippet",
			Pattern:     "nginx.ingress.kubernetes.io/stream-snippet",
			RiskLevel:   models.RiskHigh,
			Description: "Custom NGINX stream configuration for TCP/UDP",
			MigrationNote: "Stream snippets are for Layer 4 routing. Gateway API supports TCP/UDP via " +
				"TCPRoute/UDPRoute, but custom stream logic requires reimplementation.",
			SourceURL: "https://gateway-api.sigs.k8s.io/reference/spec/#tcproute",
		},
		{
			Name:        "Http Snippet",
			Pattern:     "nginx.ingress.kubernetes.io/http-snippet",
			RiskLevel:   models.RiskHigh,
			Description: "Custom NGINX http block configuration",
			MigrationNote: "HTTP snippets affect global behavior. Requires careful analysis and " +
				"potential migration to Gateway-level policies or infrastructure changes.",
			SourceURL: "https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/annotations/#configuration-snippet",
		},
	}
}

// GetRuleByPattern returns the rule that matches an annotation pattern
func GetRuleByPattern(annotationKey string) *models.AnnotationRule {
	rules := GetAnnotationRules()

	for _, rule := range rules {
		if rule.Pattern == annotationKey {
			return &rule
		}
	}

	return nil
}

// MatchAnnotations finds all rules that match the given annotations
func MatchAnnotations(annotations map[string]string) []models.AnnotationRule {
	var matchedRules []models.AnnotationRule
	rules := GetAnnotationRules()

	for annotationKey := range annotations {
		for _, rule := range rules {
			if matches, _ := regexp.MatchString(rule.Pattern, annotationKey); matches {
				matchedRules = append(matchedRules, rule)
				break // Only match each annotation once
			}
		}
	}

	return matchedRules
}

// GetUnknownNginxAnnotations identifies nginx annotations not in our rules
func GetUnknownNginxAnnotations(annotations map[string]string) []string {
	var unknown []string
	rules := GetAnnotationRules()

	// Create a set of known patterns
	knownPatterns := make(map[string]bool)
	for _, rule := range rules {
		knownPatterns[rule.Pattern] = true
	}

	for annotationKey := range annotations {
		// Check if it's an nginx annotation
		if strings.HasPrefix(annotationKey, "nginx.ingress.kubernetes.io/") {
			// Check if we have a rule for it
			if !knownPatterns[annotationKey] {
				unknown = append(unknown, annotationKey)
			}
		}
	}

	return unknown
}

// GetHighestRiskLevel determines the highest risk level from a set of rules
func GetHighestRiskLevel(rules []models.AnnotationRule) models.RiskLevel {
	if len(rules) == 0 {
		return models.RiskAuto
	}

	highestRisk := models.RiskAuto

	for _, rule := range rules {
		switch rule.RiskLevel {
		case models.RiskHigh:
			return models.RiskHigh // Highest possible, return immediately
		case models.RiskManual:
			highestRisk = models.RiskManual
		}
	}

	return highestRisk
}
